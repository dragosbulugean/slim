package util

import (
  "bytes"
  "fmt"
  "go/scanner"
  "go/types"
  "os/exec"
  "path/filepath"
  "syscall"
  "text/template"

  "go/build"
  "io"
  "net"
  "net/http"
  "os"
  "path"
  "time"

  "log"

  "github.com/fsnotify/fsnotify"
  gbuild "github.com/gopherjs/gopherjs/build"
  "github.com/gopherjs/gopherjs/compiler"
  "github.com/neelance/sourcemap"
)

func Init(webAppDirectory string) {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }
  defer watcher.Close()

  go func() {
    for {
      select {
      case event := <-watcher.Events:
        if event.Op & fsnotify.Write == fsnotify.Write {
          log.Println("modified file:", event.Name)
        }
      }
    }
  }()

  err = watcher.Add(webAppDirectory)
  if err != nil {
    log.Fatal(err)
  }
}

var currentDirectory string

func CompileAeonApp(aeonAppDirectory string) {
  go func() {
    if len(aeonAppDirectory) == 0 {
      currentDirectory = "./web"
    } else {
      currentDirectory = aeonAppDirectory
    }
    pkgObj := currentDirectory + "/web.js"
    options := &gbuild.Options{
      CreateMapFile: true,
      Minify:        true,
    }
    options.BuildTags = []string{}
    s := gbuild.NewSession(options)
    err := s.BuildDir(currentDirectory, currentDirectory, pkgObj)
    if err != nil {
      panic(err)
    }
    fmt.Println("Compiled web app successfully!")
  }()
}

func GopherJSFileServer(rootDirectory string) http.Handler {
  options := &gbuild.Options{CreateMapFile: true, Minify: true}
  options.BuildTags = []string{}
  dirs := append(filepath.SplitList(build.Default.GOPATH), build.Default.GOROOT)
  var root string

  root = rootDirectory

  sourceFiles := http.FileServer(ServeCommandFileSystem{
    serveRoot:  root,
    options:    options,
    dirs:       dirs,
    sourceMaps: make(map[string][]byte),
  })

  return sourceFiles
}

type ServeCommandFileSystem struct {
  serveRoot  string
  options    *gbuild.Options
  dirs       []string
  sourceMaps map[string][]byte
}

var cache = map[string]*fakeFile{}

func (fs ServeCommandFileSystem) Open(requestName string) (http.File, error) {

  // start := time.Now()

  name := path.Join(fs.serveRoot, requestName[1:]) // requestName[0] == '/'
  dir, file := path.Split(name)
  base := path.Base(dir) // base is parent folder name, which becomes the output file name.

  isPkg := file == base + ".js"
  isMap := file == base + ".js.map"
  isIndex := file == "index.html"

  if isPkg || isMap || isIndex {
    // If we're going to be serving our special files, make sure there's a Go command in this folder.
    s := gbuild.NewSession(fs.options)
    pkg, err := gbuild.Import(path.Dir(name), 0, s.InstallSuffix(), fs.options.BuildTags)
    if err != nil || pkg.Name != "main" {
      isPkg = false
      isMap = false
      isIndex = false
    }

    switch {
    case isPkg:
      filename := base + ".js"
      //if cached, ok := cache[filename]; ok {
      //	return cached, nil
      //}
      buf := bytes.NewBuffer(nil)
      browserErrors := bytes.NewBuffer(nil)
      exitCode := handleError(func() error {
        archive, err := s.BuildPackage(pkg)
        if err != nil {
          return err
        }

        sourceMapFilter := &compiler.SourceMapFilter{Writer: buf}
        m := &sourcemap.Map{File: base + ".js"}
        sourceMapFilter.MappingCallback = gbuild.NewMappingCallback(m, fs.options.GOROOT, fs.options.GOPATH, fs.options.MapToLocalDisk)

        deps, err := compiler.ImportDependencies(archive, s.BuildImportPath)
        if err != nil {
          return err
        }
        if err := compiler.WriteProgramCode(deps, sourceMapFilter); err != nil {
          return err
        }

        mapBuf := bytes.NewBuffer(nil)
        m.WriteTo(mapBuf)
        buf.WriteString("//# sourceMappingURL=" + base + ".js.map\n")
        fs.sourceMaps[name + ".map"] = mapBuf.Bytes()

        return nil
      }, fs.options, browserErrors)
      if exitCode != 0 {
        buf = browserErrors
      }
      // elapsed := time.Since(start)
      // log.Println("%s took %s", file, elapsed)
      cache[filename] = newFakeFile(filename, buf.Bytes())
      return cache[filename], nil

    case isMap:
      if content, ok := fs.sourceMaps[name]; ok {
        filename := base + ".js.map"
        cache[filename] = newFakeFile(base + ".js.map", content)
        return cache[filename], nil
      }
    }
  }

  for _, d := range fs.dirs {
    dir := http.Dir(filepath.Join(d, "src"))
    f, err := dir.Open(name)
    if err == nil {
      return f, nil
    }

    // source maps are served outside of serveRoot
    f, err = dir.Open(requestName)
    if err == nil {
      return f, nil
    }
  }

  if isIndex {
    // If there was no index.html file in any dirs, supply our own.
    return newFakeFile("index.html", []byte(`<html><head><meta charset="utf-8"><script src="` + base + `.js"></script></head><body></body></html>`)), nil
  }

  return nil, os.ErrNotExist
}

type fakeFile struct {
  name string
  size int
  io.ReadSeeker
}

func newFakeFile(name string, content []byte) *fakeFile {
  return &fakeFile{name: name, size: len(content), ReadSeeker: bytes.NewReader(content)}
}

func (f *fakeFile) Close() error {
  return nil
}

func (f *fakeFile) Readdir(count int) ([]os.FileInfo, error) {
  return nil, os.ErrInvalid
}

func (f *fakeFile) Stat() (os.FileInfo, error) {
  return f, nil
}

func (f *fakeFile) Name() string {
  return f.name
}

func (f *fakeFile) Size() int64 {
  return int64(f.size)
}

func (f *fakeFile) Mode() os.FileMode {
  return 0
}

func (f *fakeFile) ModTime() time.Time {
  return time.Time{}
}

func (f *fakeFile) IsDir() bool {
  return false
}

func (f *fakeFile) Sys() interface{} {
  return nil
}

// If browserErrors is non-nil, errors are written for presentation in browser.
func handleError(f func() error, options *gbuild.Options, browserErrors *bytes.Buffer) int {
  switch err := f().(type) {
  case nil:
    return 0
  case compiler.ErrorList:
    for _, entry := range err {
      printError(entry, options, browserErrors)
    }
    return 1
  case *exec.ExitError:
    return err.Sys().(syscall.WaitStatus).ExitStatus()
  default:
    printError(err, options, browserErrors)
    return 1
  }
}

// sprintError returns an annotated error string without trailing newline.
func sprintError(err error) string {
  makeRel := func(name string) string {
    if relname, err := filepath.Rel(currentDirectory, name); err == nil {
      return relname
    }
    return name
  }

  switch e := err.(type) {
  case *scanner.Error:
    return fmt.Sprintf("%s:%d:%d: %s", makeRel(e.Pos.Filename), e.Pos.Line, e.Pos.Column, e.Msg)
  case types.Error:
    pos := e.Fset.Position(e.Pos)
    return fmt.Sprintf("%s:%d:%d: %s", makeRel(pos.Filename), pos.Line, pos.Column, e.Msg)
  default:
    return fmt.Sprintf("%s", e)
  }
}

// printError prints err to Stderr with options. If browserErrors is non-nil, errors are also written for presentation in browser.
func printError(err error, options *gbuild.Options, browserErrors *bytes.Buffer) {
  e := sprintError(err)
  options.PrintError("%s\n", e)
  if browserErrors != nil {
    fmt.Fprintln(browserErrors, `console.error("` + template.JSEscapeString(e) + `");`)
  }
}

type tcpKeepAliveListener struct {
  *net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
  tc, err := ln.AcceptTCP()
  if err != nil {
    return
  }
  tc.SetKeepAlive(true)
  tc.SetKeepAlivePeriod(3 * time.Minute)
  return tc, nil
}
