package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	handler "github.com/simia-tech/graphql-fasthttp-handler"
)

func TestContextPropagated(t *testing.T) {
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Context.Value("name"), nil
				},
			},
		},
	})
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{"name": nil},
	}
	queryString := `query={name}`

	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost")
	req.Header.SetMethod(fasthttp.MethodGet)
	req.URI().SetPath("/graphql")
	req.URI().SetQueryString(queryString)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
	})

	if err := serve(h.ServeHTTP, req, resp); err != nil {
		t.Fatal(err)
	}

	result := decodeResponse(t, resp)
	if code := resp.StatusCode(); code != fasthttp.StatusOK {
		t.Fatalf("unexpected server response %v", code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_BasicQuery_Pretty(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}
	queryString := `query=query HeroNameQuery { hero { name } }&operationName=HeroNameQuery`

	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost")
	req.Header.SetMethod(fasthttp.MethodGet)
	req.URI().SetPath("/graphql")
	req.URI().SetQueryString(queryString)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	callbackCalled := false
	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
		Pretty: true,
		ResultCallbackFn: func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte) {
			callbackCalled = true
			if params.OperationName != "HeroNameQuery" {
				t.Fatalf("OperationName passed to callback was not HeroNameQuery: %v", params.OperationName)
			}

			if result.HasErrors() {
				t.Fatalf("unexpected graphql result errors")
			}
		},
	})

	if err := serve(h.ServeHTTP, req, resp); err != nil {
		t.Fatal(err)
	}

	result := decodeResponse(t, resp)
	if code := resp.StatusCode(); code != http.StatusOK {
		t.Fatalf("unexpected server response %v", code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
	if !callbackCalled {
		t.Fatalf("ResultCallbackFn was not called when it should have been")
	}
}

func TestHandler_BasicQuery_Ugly(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}
	queryString := `query=query HeroNameQuery { hero { name } }`

	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost")
	req.Header.SetMethod(fasthttp.MethodGet)
	req.URI().SetPath("/graphql")
	req.URI().SetQueryString(queryString)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
		Pretty: false,
	})

	if err := serve(h.ServeHTTP, req, resp); err != nil {
		t.Fatal(err)
	}

	result := decodeResponse(t, resp)
	if code := resp.StatusCode(); code != http.StatusOK {
		t.Fatalf("unexpected server response %v", code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Params_NilParams(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if str, ok := r.(string); ok {
				if str != "undefined GraphQL schema" {
					t.Fatalf("unexpected error, got %v", r)
				}
				// test passed
				return
			}
			t.Fatalf("unexpected error, got %v", r)

		}
		t.Fatalf("expected to panic, did not panic")
	}()
	_ = handler.New(nil)

}

func TestHandler_BasicQuery_WithRootObjFn(t *testing.T) {
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					rv := p.Info.RootValue.(map[string]interface{})
					return rv["rootValue"], nil
				},
			},
		},
	})
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{
			"name": "foo",
		},
	}
	queryString := `query={name}`

	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost")
	req.Header.SetMethod(fasthttp.MethodGet)
	req.URI().SetPath("/graphql")
	req.URI().SetQueryString(queryString)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
		RootObjectFn: func(reqCtx *fasthttp.RequestCtx) map[string]interface{} {
			return map[string]interface{}{"rootValue": "foo"}
		},
	})

	if err := serve(h.ServeHTTP, req, resp); err != nil {
		t.Fatal(err)
	}

	result := decodeResponse(t, resp)
	if code := resp.StatusCode(); code != http.StatusOK {
		t.Fatalf("unexpected server response %v", code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

type customError struct {
	message string
}

func (e customError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

func TestHandler_BasicQuery_WithFormatErrorFn(t *testing.T) {
	resolverError := customError{message: "resolver error"}
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, resolverError
				},
			},
		},
	})
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
	if err != nil {
		t.Fatal(err)
	}

	customFormattedError := gqlerrors.FormattedError{
		Message: resolverError.Error(),
		Locations: []location.SourceLocation{
			location.SourceLocation{
				Line:   1,
				Column: 2,
			},
		},
		Path: []interface{}{"name"},
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{
			"name": nil,
		},
		Errors: []gqlerrors.FormattedError{customFormattedError},
	}

	queryString := `query={name}`

	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost")
	req.Header.SetMethod(fasthttp.MethodGet)
	req.URI().SetPath("/graphql")
	req.URI().SetQueryString(queryString)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	formatErrorFnCalled := false
	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
		FormatErrorFn: func(err error) gqlerrors.FormattedError {
			formatErrorFnCalled = true
			var formatted gqlerrors.FormattedError
			switch err := err.(type) {
			case *gqlerrors.Error:
				formatted = gqlerrors.FormatError(err)
			default:
				t.Fatalf("unexpected error type: %v", reflect.TypeOf(err))
			}
			return formatted
		},
	})

	if err := serve(h.ServeHTTP, req, resp); err != nil {
		t.Fatal(err)
	}

	result := decodeResponse(t, resp)
	if code := resp.StatusCode(); code != http.StatusOK {
		t.Fatalf("unexpected server response %v", code)
	}
	if !formatErrorFnCalled {
		t.Fatalf("FormatErrorFn was not called when it should have been")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func decodeResponse(t *testing.T, response *fasthttp.Response) *graphql.Result {
	var target graphql.Result
	if err := json.Unmarshal(response.Body(), &target); err != nil {
		t.Fatalf("DecodeResponseToType(): %v \n%s", err.Error(), response.Body())
	}
	return &target
}

func serve(handler fasthttp.RequestHandler, req *fasthttp.Request, res *fasthttp.Response) error {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(err)
		}
	}()

	client := fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}
