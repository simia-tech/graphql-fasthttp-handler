package handler_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	"github.com/valyala/fasthttp"

	handler "github.com/simia-tech/graphql-fasthttp-handler"
)

func TestRequestOptions(t *testing.T) {
	testGenFn := func(method string) func(string, string, string, *handler.RequestOptions) func(*testing.T) {
		return func(query, contentType, body string, expected *handler.RequestOptions) func(*testing.T) {
			return func(t *testing.T) {
				req := &fasthttp.Request{}
				req.Header.SetMethod(method)
				req.URI().SetPath("/graphql")
				if len(query) > 0 {
					req.URI().SetQueryString(query)
				}
				if len(contentType) > 0 {
					req.Header.SetContentType(contentType)
				}
				if len(body) > 0 {
					io.Copy(req.BodyWriter(), strings.NewReader(body))
				}

				result := handler.NewRequestOptions(req)

				if !reflect.DeepEqual(result, expected) {
					t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
				}
			}
		}
	}

	t.Run("GET", func(t *testing.T) {
		testFn := testGenFn(fasthttp.MethodGet)

		t.Run("BasicQueryString", testFn(
			"query=query RebelsShipsQuery { rebels { name } }",
			"",
			"",
			&handler.RequestOptions{
				Query:     "query RebelsShipsQuery { rebels { name } }",
				Variables: make(map[string]interface{}),
			},
		))
		t.Run("ContentTypeApplicationGraphQL", testFn(
			"",
			"application/graphql",
			"query=query RebelsShipsQuery { rebels { name } }",
			&handler.RequestOptions{},
		))
		t.Run("ContentTypeApplicationJSON", testFn(
			"",
			"application/json",
			`{ "query": "query RebelsShipsQuery { rebels { name } }" }`,
			&handler.RequestOptions{},
		))
	})

	t.Run("POST", func(t *testing.T) {
		testFn := testGenFn(fasthttp.MethodPost)

		t.Run("BasicQueryStringWithNoBody", testFn(
			"query=query RebelsShipsQuery { rebels { name } }",
			"",
			"",
			&handler.RequestOptions{
				Query:     "query RebelsShipsQuery { rebels { name } }",
				Variables: make(map[string]interface{}),
			},
		))
		t.Run("BasicQueryStringWithNoBody", testFn(
			"",
			"application/graphql",
			"query RebelsShipsQuery { rebels { name } }",
			&handler.RequestOptions{
				Query: "query RebelsShipsQuery { rebels { name } }",
			},
		))
	})
}

// func TestRequestOptions_POST_ContentTypeApplicationGraphQL_WithNonGraphQLQueryContent(t *testing.T) {
// 	body := []byte(`not a graphql query`)
// 	expected := &RequestOptions{
// 		Query: "not a graphql query",
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
// 	req.Header.Add("Content-Type", "application/graphql")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationGraphQL_EmptyBody(t *testing.T) {
// 	body := []byte(``)
// 	expected := &RequestOptions{
// 		Query: "",
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
// 	req.Header.Add("Content-Type", "application/graphql")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationGraphQL_NilBody(t *testing.T) {
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", nil)
// 	req.Header.Add("Content-Type", "application/graphql")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_POST_ContentTypeApplicationJSON(t *testing.T) {
// 	body := `
// 	{
// 		"query": "query RebelsShipsQuery { rebels { name } }"
// 	}`
// 	expected := &RequestOptions{
// 		Query: "query RebelsShipsQuery { rebels { name } }",
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_GET_WithVariablesAsObject(t *testing.T) {
// 	variables := url.QueryEscape(`{ "a": 1, "b": "2" }`)
// 	query := url.QueryEscape("query RebelsShipsQuery { rebels { name } }")
// 	queryString := fmt.Sprintf("query=%s&variables=%s", query, variables)
// 	expected := &RequestOptions{
// 		Query: "query RebelsShipsQuery { rebels { name } }",
// 		Variables: map[string]interface{}{
// 			"a": float64(1),
// 			"b": "2",
// 		},
// 	}

// 	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_POST_ContentTypeApplicationJSON_WithVariablesAsObject(t *testing.T) {
// 	body := `
// 	{
// 		"query": "query RebelsShipsQuery { rebels { name } }",
// 		"variables": { "a": 1, "b": "2" }
// 	}`
// 	expected := &RequestOptions{
// 		Query: "query RebelsShipsQuery { rebels { name } }",
// 		Variables: map[string]interface{}{
// 			"a": float64(1),
// 			"b": "2",
// 		},
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationJSON_WithVariablesAsString(t *testing.T) {
// 	body := `
// 	{
// 		"query": "query RebelsShipsQuery { rebels { name } }",
// 		"variables": "{ \"a\": 1, \"b\": \"2\" }"
// 	}`
// 	expected := &RequestOptions{
// 		Query: "query RebelsShipsQuery { rebels { name } }",
// 		Variables: map[string]interface{}{
// 			"a": float64(1),
// 			"b": "2",
// 		},
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationJSON_WithInvalidJSON(t *testing.T) {
// 	body := `INVALIDJSON{}`
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationJSON_WithNilBody(t *testing.T) {
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", nil)
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_POST_ContentTypeApplicationUrlEncoded(t *testing.T) {
// 	data := url.Values{}
// 	data.Add("query", "query RebelsShipsQuery { rebels { name } }")

// 	expected := &RequestOptions{
// 		Query:     "query RebelsShipsQuery { rebels { name } }",
// 		Variables: make(map[string]interface{}),
// 	}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(data.Encode()))
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationUrlEncoded_WithInvalidData(t *testing.T) {
// 	data := "Invalid Data"

// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(data))
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_POST_ContentTypeApplicationUrlEncoded_WithNilBody(t *testing.T) {

// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", nil)
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_PUT_BasicQueryString(t *testing.T) {
// 	queryString := "query=query RebelsShipsQuery { rebels { name } }"
// 	expected := &RequestOptions{
// 		Query:     "query RebelsShipsQuery { rebels { name } }",
// 		Variables: make(map[string]interface{}),
// 	}

// 	req, _ := http.NewRequest("PUT", fmt.Sprintf("/graphql?%v", queryString), nil)
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_PUT_ContentTypeApplicationGraphQL(t *testing.T) {
// 	body := []byte(`query RebelsShipsQuery { rebels { name } }`)
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("PUT", "/graphql", bytes.NewBuffer(body))
// 	req.Header.Add("Content-Type", "application/graphql")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_PUT_ContentTypeApplicationJSON(t *testing.T) {
// 	body := `
// 	{
// 		"query": "query RebelsShipsQuery { rebels { name } }"
// 	}`
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("PUT", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_PUT_ContentTypeApplicationUrlEncoded(t *testing.T) {
// 	data := url.Values{}
// 	data.Add("query", "query RebelsShipsQuery { rebels { name } }")

// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("PUT", "/graphql", bytes.NewBufferString(data.Encode()))
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_DELETE_BasicQueryString(t *testing.T) {
// 	queryString := "query=query RebelsShipsQuery { rebels { name } }"
// 	expected := &RequestOptions{
// 		Query:     "query RebelsShipsQuery { rebels { name } }",
// 		Variables: make(map[string]interface{}),
// 	}

// 	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/graphql?%v", queryString), nil)
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_DELETE_ContentTypeApplicationGraphQL(t *testing.T) {
// 	body := []byte(`query RebelsShipsQuery { rebels { name } }`)
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("DELETE", "/graphql", bytes.NewBuffer(body))
// 	req.Header.Add("Content-Type", "application/graphql")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_DELETE_ContentTypeApplicationJSON(t *testing.T) {
// 	body := `
// 	{
// 		"query": "query RebelsShipsQuery { rebels { name } }"
// 	}`
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("DELETE", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/json")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
// func TestRequestOptions_DELETE_ContentTypeApplicationUrlEncoded(t *testing.T) {
// 	data := url.Values{}
// 	data.Add("query", "query RebelsShipsQuery { rebels { name } }")

// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("DELETE", "/graphql", bytes.NewBufferString(data.Encode()))
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }

// func TestRequestOptions_POST_UnsupportedContentType(t *testing.T) {
// 	body := `<xml>query{}</xml>`
// 	expected := &RequestOptions{}

// 	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
// 	req.Header.Add("Content-Type", "application/xml")
// 	result := NewRequestOptions(req)

// 	if !reflect.DeepEqual(result, expected) {
// 		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
// 	}
// }
