package api

import (
  "net/http"
  "encoding/json"
  "strings"
  "bitbucket.org/dragosbulugean/aiurlabs/slim/shared"
)

func CreateSlimLink(url string) shared.CreateSlimLinkResponse {
  marshalled, _ := json.Marshal(shared.CreateSlimLinkRequest{URL: url})
  response, err := http.Post(shared.Routes.Slim, "application/json", strings.NewReader(string(marshalled)))
  if err != nil {
    println(err)
  }
  defer response.Body.Close()
  createSlimLinkResponse := shared.CreateSlimLinkResponse{}
  json.NewDecoder(response.Body).Decode(&createSlimLinkResponse)
  return createSlimLinkResponse
}

