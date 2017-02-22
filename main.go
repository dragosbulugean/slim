package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/dragosbulugean/aiurlabs/slim/util"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
  "time"
  "bitbucket.org/dragosbulugean/aiurlabs/slim/shared"
  "github.com/asdine/storm"

  "io/ioutil"
  "github.com/dchest/uniuri"
)

var (
	aeonAppPath = "bitbucket.org/dragosbulugean/aiurlabs/slim/web"
  db *storm.DB
)

//Model
type UrlModel struct {
  ID int `storm:"id,increment"`
  SlimURL string `storm:"index"`
  URL string
}

func main() {

  db, _ = storm.Open("slim.db")

  r := mux.NewRouter()
	r.HandleFunc(shared.Routes.Slim, slimURL).Methods("POST")
  r.HandleFunc(shared.Routes.GoTo, goToSlimURL).Methods("GET")
  r.PathPrefix("/").Handler(handlers.CompressHandler(util.GopherJSFileServer(aeonAppPath)))
  loggedRouter := handlers.LoggingHandler(os.Stdout, r)
  http.Handle("/", loggedRouter)

  srv := &http.Server{
    Handler:      r,
    Addr:         "127.0.0.1:3000",
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }
  log.Fatal(srv.ListenAndServe())
}

func slimURL(w http.ResponseWriter, r *http.Request) {
  body, _ := ioutil.ReadAll(r.Body)
  createSlimLinkRequest := shared.CreateSlimLinkRequest{}
  err := json.Unmarshal(body, &createSlimLinkRequest)
  if err != nil {
    fmt.Print(err)
  }
  slimURL := UrlModel {
    URL: createSlimLinkRequest.URL,
    SlimURL: uniuri.NewLen(6),
  }
  err = db.Save(&slimURL)
  if err != nil {
    fmt.Print(err)
  }
	resp := &shared.CreateSlimLinkResponse{Status: http.StatusOK, URL: slimURL.SlimURL}
	respJSON, _ := json.Marshal(resp)
	fmt.Fprint(w, string(respJSON))
	return
}

func goToSlimURL(w http.ResponseWriter, r *http.Request) {
  url := mux.Vars(r)["url"]
  var slimURL UrlModel
  _ = db.One("SlimURL", url, &slimURL)
  http.Redirect(w, r, slimURL.URL, http.StatusSeeOther)
}
