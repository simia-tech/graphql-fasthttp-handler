package handler

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/valyala/fasthttp"
)

type playgroundData struct {
	PlaygroundVersion    string
	Endpoint             string
	SubscriptionEndpoint string
	SetTitle             bool
	Path                 string
}

// renderPlayground renders the Playground GUI
func renderPlayground(reqCtx *fasthttp.RequestCtx) {
	t := template.New("Playground")
	t, err := t.Parse(graphcoolPlaygroundTemplate)
	if err != nil {
		reqCtx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	reqCtx.Response.Header.SetContentType("text/html; charset=utf-8")

	d := playgroundData{
		PlaygroundVersion:    graphcoolPlaygroundVersion,
		Endpoint:             string(reqCtx.Request.URI().Path()),
		SubscriptionEndpoint: fmt.Sprintf("ws://%s/subscriptions", reqCtx.Request.Header.Host()),
		SetTitle:             true,
		Path:                 strings.TrimPrefix(string(reqCtx.Path()), "/"),
	}
	err = t.ExecuteTemplate(reqCtx, "index", d)
	if err != nil {
		reqCtx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	return
}

const graphcoolPlaygroundVersion = "1.5.2"

const graphcoolPlaygroundTemplate = `
{{ define "index" }}
<!--
The request to this GraphQL server provided the header "Accept: text/html"
and as a result has been presented Playground - an in-browser IDE for
exploring GraphQL.

If you wish to receive JSON, provide the header "Accept: application/json" or
add "&raw" to the end of the URL within a browser.
-->
<!DOCTYPE html>
<html>

<head>
  <meta charset=utf-8/>
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>GraphQL Playground</title>
  <link rel="stylesheet" href="{{ .Path }}/static/playground/index.css" />
  <link rel="shortcut icon" href="{{ .Path }}/static/playground/favicon.png" />
  <script src="{{ .Path }}/static/playground/middleware.js"></script>
</head>

<body>
  <div id="root">
    <style>
      body {
        background-color: rgb(23, 42, 58);
        font-family: Open Sans, sans-serif;
        height: 90vh;
      }
      #root {
        height: 100%;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .loading {
        font-size: 32px;
        font-weight: 200;
        color: rgba(255, 255, 255, .6);
        margin-left: 20px;
      }
      img {
        width: 78px;
        height: 78px;
      }
      .title {
        font-weight: 400;
      }
    </style>
    <img src='{{ .Path }}/static/playground/logo.png' alt=''>
    <div class="loading"> Loading
      <span class="title">GraphQL Playground</span>
    </div>
  </div>
  <script>window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        // options as 'endpoint' belong here
        endpoint: {{ .Endpoint }},
        subscriptionEndpoint: {{ .SubscriptionEndpoint }},
        setTitle: {{ .SetTitle }}
      })
    })</script>
</body>

</html>
{{ end }}
`
