package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/graphql-go/graphql"
	"github.com/valyala/fasthttp"

	"context"

	"github.com/graphql-go/graphql/gqlerrors"
)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeGraphQL        = "application/graphql"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

type ResultCallbackFn func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte)

type Handler struct {
	Schema           *graphql.Schema
	pretty           bool
	graphiql         bool
	playground       bool
	rootObjectFn     RootObjectFn
	resultCallbackFn ResultCallbackFn
	formatErrorFn    func(err error) gqlerrors.FormattedError
}

type RequestOptions struct {
	Query         string                 `json:"query" url:"query" schema:"query"`
	Variables     map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
	OperationName string                 `json:"operationName" url:"operationName" schema:"operationName"`
}

// a workaround for getting`variables` as a JSON string
type requestOptionsCompatibility struct {
	Query         string `json:"query" url:"query" schema:"query"`
	Variables     string `json:"variables" url:"variables" schema:"variables"`
	OperationName string `json:"operationName" url:"operationName" schema:"operationName"`
}

func getFromForm(args *fasthttp.Args) *RequestOptions {
	query := args.Peek("query")
	if len(query) > 0 {
		// get variables map
		variables := make(map[string]interface{})
		variablesBytes := args.Peek("variables")
		json.Unmarshal(variablesBytes, &variables)

		return &RequestOptions{
			Query:         string(query),
			Variables:     variables,
			OperationName: string(args.Peek("operationName")),
		}
	}

	return nil
}

// NewRequestOptions Parses a http.Request into GraphQL request options struct
func NewRequestOptions(r *fasthttp.Request) *RequestOptions {
	if reqOpt := getFromForm(r.URI().QueryArgs()); reqOpt != nil {
		return reqOpt
	}

	if !r.Header.IsPost() {
		return &RequestOptions{}
	}

	if r.Body() == nil {
		return &RequestOptions{}
	}

	// TODO: improve Content-Type handling
	contentTypeStr := string(r.Header.ContentType())
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case ContentTypeGraphQL:
		return &RequestOptions{
			Query: string(r.Body()),
		}
	case ContentTypeFormURLEncoded:
		// if err := r.ParseForm(); err != nil {
		// 	return &RequestOptions{}
		// }

		if reqOpt := getFromForm(r.PostArgs()); reqOpt != nil {
			return reqOpt
		}

		return &RequestOptions{}

	case ContentTypeJSON:
		fallthrough
	default:
		var opts RequestOptions
		if err := json.Unmarshal(r.Body(), &opts); err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var optsCompatible requestOptionsCompatibility
			json.Unmarshal(r.Body(), &optsCompatible)
			json.Unmarshal([]byte(optsCompatible.Variables), &opts.Variables)
		}
		return &opts
	}
}

// ServeHTTP provides an entrypoint into executing graphQL queries.
func (h *Handler) ServeHTTP(reqCtx *fasthttp.RequestCtx) {
	// get query
	opts := NewRequestOptions(&reqCtx.Request)

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        reqCtx,
	}
	if h.rootObjectFn != nil {
		params.RootObject = h.rootObjectFn(reqCtx)
	}
	result := graphql.Do(params)

	if formatErrorFn := h.formatErrorFn; formatErrorFn != nil && len(result.Errors) > 0 {
		formatted := make([]gqlerrors.FormattedError, len(result.Errors))
		for i, formattedError := range result.Errors {
			formatted[i] = formatErrorFn(formattedError.OriginalError())
		}
		result.Errors = formatted
	}

	if h.graphiql {
		acceptHeader := string(reqCtx.Request.Header.Peek("Accept"))
		raw := reqCtx.Request.URI().QueryArgs().Has("raw")
		if !raw && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			renderGraphiQL(reqCtx, params)
			return
		}
	}

	if h.playground {
		acceptHeader := string(reqCtx.Request.Header.Peek("Accept"))
		raw := reqCtx.Request.URI().QueryArgs().Has("raw")
		if !raw && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			renderPlayground(reqCtx)
			return
		}
	}

	if bytes.Equal(reqCtx.Request.Header.Method(), []byte(fasthttp.MethodGet)) && bytes.Contains(reqCtx.URI().Path(), []byte("/static/")) {
		serveStatic(reqCtx)
		return
	}

	// use proper JSON Header
	reqCtx.Response.Header.SetContentType("application/json; charset=utf-8")

	var buff []byte
	if h.pretty {
		reqCtx.SetStatusCode(fasthttp.StatusOK)
		buff, _ = json.MarshalIndent(result, "", "\t")
		reqCtx.Write(buff)
	} else {
		reqCtx.SetStatusCode(fasthttp.StatusOK)
		buff, _ = json.Marshal(result)
		reqCtx.Write(buff)
	}

	if h.resultCallbackFn != nil {
		h.resultCallbackFn(reqCtx, &params, result, buff)
	}
}

// RootObjectFn allows a user to generate a RootObject per request
type RootObjectFn func(reqCtx *fasthttp.RequestCtx) map[string]interface{}

type Config struct {
	Schema           *graphql.Schema
	Pretty           bool
	GraphiQL         bool
	Playground       bool
	RootObjectFn     RootObjectFn
	ResultCallbackFn ResultCallbackFn
	FormatErrorFn    func(err error) gqlerrors.FormattedError
}

func NewConfig() *Config {
	return &Config{
		Schema:     nil,
		Pretty:     true,
		GraphiQL:   true,
		Playground: false,
	}
}

func New(p *Config) *Handler {
	if p == nil {
		p = NewConfig()
	}

	if p.Schema == nil {
		panic("undefined GraphQL schema")
	}

	return &Handler{
		Schema:           p.Schema,
		pretty:           p.Pretty,
		graphiql:         p.GraphiQL,
		playground:       p.Playground,
		rootObjectFn:     p.RootObjectFn,
		resultCallbackFn: p.ResultCallbackFn,
		formatErrorFn:    p.FormatErrorFn,
	}
}

var staticBox = packr.New("graphql-web", "./static")

func serveStatic(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	path = path[strings.Index(path, "/static/")+8:]

	content, err := staticBox.Find(path)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s - 404 bot found", path)
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("err %v", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = http.DetectContentType(content[:1024])
	}
	ctx.Response.Header.SetContentType(contentType)

	ctx.Response.SetBody(content)
}
