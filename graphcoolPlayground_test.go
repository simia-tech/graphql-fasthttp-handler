package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	handler "github.com/simia-tech/graphql-fasthttp-handler"
	"github.com/valyala/fasthttp"
)

func TestRenderPlayground(t *testing.T) {
	cases := map[string]struct {
		playgroundEnabled    bool
		accept               string
		url                  string
		expectedStatusCode   int
		expectedContentType  string
		expectedBodyContains string
	}{
		"renders Playground": {
			playgroundEnabled:    true,
			accept:               "text/html",
			expectedStatusCode:   http.StatusOK,
			expectedContentType:  "text/html; charset=utf-8",
			expectedBodyContains: "<!DOCTYPE html>",
		},
		"doesn't render Playground if turned off": {
			playgroundEnabled:   false,
			accept:              "text/html",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render Playground if Content-Type application/json is present": {
			playgroundEnabled:   true,
			accept:              "application/json,text/html",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render Playground if Content-Type text/html is not present": {
			playgroundEnabled:   true,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render Playground if 'raw' query is present": {
			playgroundEnabled:   true,
			accept:              "text/html",
			url:                 "?raw",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
	}

	for tcID, tc := range cases {
		t.Run(tcID, func(t *testing.T) {
			req := fasthttp.AcquireRequest()
			req.Header.SetHost("localhost")
			req.Header.SetMethod(fasthttp.MethodGet)
			req.Header.Set("Accept", tc.accept)
			req.URI().SetPath(tc.url)
			defer fasthttp.ReleaseRequest(req)

			resp := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(resp)

			h := handler.New(&handler.Config{
				Schema:     &testutil.StarWarsSchema,
				GraphiQL:   false,
				Playground: tc.playgroundEnabled,
			})

			if err := serve(h.ServeHTTP, req, resp); err != nil {
				t.Fatal(err)
			}

			if statusCode := resp.StatusCode(); statusCode != tc.expectedStatusCode {
				t.Fatalf("%s: wrong status code, expected %v, got %v", tcID, tc.expectedStatusCode, statusCode)
			}

			if contentType := string(resp.Header.ContentType()); contentType != tc.expectedContentType {
				t.Fatalf("%s: wrong content type, expected %s, got %s", tcID, tc.expectedContentType, contentType)
			}

			if body := string(resp.Body()); !strings.Contains(body, tc.expectedBodyContains) {
				t.Fatalf("%s: wrong body, expected %s to contain %s", tcID, body, tc.expectedBodyContains)
			}
		})
	}
}
