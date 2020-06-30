package deploy

import (
	"avanoo_cd/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/gorilla/mux"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type JSONTestRequest struct {
	method      string
	url         string
	headers     map[string][]string
	routeVars   map[string]string
	queryParams map[string]string
	body        io.Reader
	handlerFunc http.HandlerFunc
}

func cleanRedis() {
	utils.RedisClient.FlushDB(context.Background())
}

func createTestRequest(requestParams *JSONTestRequest) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(requestParams.method, requestParams.url, requestParams.body)
	addTestRequestHeaders(req, requestParams.headers)
	req = addRouterVars(req, requestParams.routeVars)
	addQueryParams(req, requestParams.queryParams)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(requestParams.handlerFunc)
	handler.ServeHTTP(rr, req)

	return rr
}

func addTestRequestHeaders(req *http.Request, headers map[string][]string) {
	if headers == nil {
		return
	}

	for key, value := range headers {
		for index := range value {
			req.Header.Add(key, value[index])
		}
	}
}

func addQueryParams(req *http.Request, vars map[string]string) {
	if vars == nil {
		return
	}

	q := req.URL.Query()
	for k, v := range vars {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
}

func addRouterVars(req *http.Request, vars map[string]string) *http.Request {
	if vars == nil {
		return req
	}

	return mux.SetURLVars(req, vars)
}

func assertStatusCode(responseStatus int, expectedStatus int) error {
	if responseStatus != expectedStatus {
		return fmt.Errorf("Wrong status code: got %v want %v\n", responseStatus, expectedStatus)
	}

	return nil
}

func decodeTestJSONResponse(body *bytes.Buffer, response interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	error := decoder.Decode(response)
	return error
}

func successTest(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func encodeMessage(message interface{}) io.Reader {
	encodedMessage, _ := json.Marshal(message)
	return bytes.NewReader(encodedMessage)
}

func testEqualModel(actual interface{}, expected interface{}, model interface{}, ignoreFields ...string) string {
	return cmp.Diff(actual, expected,
		cmpopts.IgnoreUnexported(model),
		cmpopts.IgnoreFields(model, ignoreFields...),
		cmpopts.EquateEmpty())
}

func getSearchValueWithDefault(searchValue string, model interface{}, field string) string {
	if searchValue != "" {
		return searchValue
	}

	if model != nil {
		fieldValue := reflect.ValueOf(model).Elem().FieldByName(field)
		if fieldValue.String() != "" {
			return fieldValue.Interface().(string)
		}
	}

	return ""
}

func testEqualObject(actual interface{}, expected interface{}) string {
	return cmp.Diff(actual, expected,
		cmpopts.EquateEmpty())
}