package main

import (
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/aeon"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/dom"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/props"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/recipes"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/router"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/std"
	aeonStyles "bitbucket.org/dragosbulugean/aiurlabs/aeon/styles"
	"bitbucket.org/dragosbulugean/aiurlabs/aeon/util/console"
	"bitbucket.org/dragosbulugean/aiurlabs/slim/web/styles"
	"bitbucket.org/dragosbulugean/aiurlabs/slim/web/state"
  "bitbucket.org/dragosbulugean/aiurlabs/slim/web/api"
)

var (
	App aeon.App
	State = new(state.State)
)

func main() {

	console.Time("RenderApp")

	appTitle := std.NewText()
	appTitle.Text = "Slim is an awesome URL shortener"
	appTitle.Class = styles.Title.Class

	urlInput := std.NewTextInput()
	urlInput.Class = styles.URLInput.Class
	urlInput.Placeholder = "Your URL"
  urlInput.Value = "http://"
	urlInput.Props.OnTextChange = func(ev *dom.Event, el *dom.HTMLInputElement) {
    urlInput.Value = el.Value
		State.UrlToBeSlimmed = el.Value
	}

  slimLink := std.NewLink()
  slimLink.InPlace = false
  slimLink.Class = styles.SlimLink.Class

	button := std.NewButton()
	button.Label = "Go slim!"
	button.Class = recipes.AeonButton.Class
	button.Props.OnMouseClick = func(event *dom.MouseEvent) {
    go func() {
      response := api.CreateSlimLink(State.UrlToBeSlimmed)
      slimLink.Text = "slim.io:3000/go/" + response.URL
      slimLink.URL = "http://" + slimLink.Text
      println("Created URL:", slimLink.Text)
      App.Render()
    }()
  }

	App = aeon.NewApp("app",
    std.Container(styles.AppContainer.Class, props.New(),
      appTitle,
      urlInput,
      button,
      slimLink,
    ),
	)

  aeonStyles.IncludeStandardCSS()
  aeonStyles.Render()

	App.Router.
		Handle("/", func(c *router.Context) {
			App.Render()
		}).
    Handle("/auth", func(c *router.Context) {
      App.Render()
    }).
		Start()

	console.TimeEnd("RenderApp")

}
