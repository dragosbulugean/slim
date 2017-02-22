package shared

type CreateSlimLinkRequest struct {
  URL string
}

type CreateSlimLinkResponse struct {
  Status int    `json:"status"`
  URL    string `json:"url"`
}

type routes struct {
  Slim string
  GoTo string
}

var Routes = routes{
  Slim: "/slim",
  GoTo: "/go/{url}",
}

