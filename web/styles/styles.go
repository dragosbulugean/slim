package styles

import (
  . "bitbucket.org/dragosbulugean/aiurlabs/aeon/styles"
)

var (
  AppContainer = New(
    "appContainer",
    map[string]string{
      S.Width:        "flex",
      S.FlexDirection:  "column",
      S.JustifyContent: "center",
      S.AlignItems:     "center",
      S.Height:         "100%",
    },
  )
  Title = New(
    "appTitle",
    map[string]string{
      S.FontSize: "140%",
    },
  )
  URLInput = New(
    "aeonTextInput",
    map[string]string{
      S.Width: "200px",
      S.MaxWidth: "200px",
      S.Padding: "10px",
      S.Margin: "20px",
    },
  )
  SlimLink = New(
    "slimLink",
    map[string]string{
      S.FontSize: "200%",
      S.PaddingTop: "20px",
    },
  )
)
