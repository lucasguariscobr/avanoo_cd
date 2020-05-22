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

// func fetchFields(obj interface{}) []string {
// 	elem := getElemObject(obj)
// 	fields := []string{}
// 	for i := 0; i < elem.NumField(); i++ {
// 		json_field := getJSONFieldName(&elem, i)
// 		fields = append(fields, json_field)
// 	}
// 	return fields
// }

// func getElemObject(obj interface{}) reflect.Value {
// 	rv := reflect.ValueOf(obj)
// 	var elem reflect.Value
// 	if rv.Type().Kind() == reflect.Slice {
// 		elem = rv.Index(0)
// 	} else {
// 		elem = rv.Elem()
// 	}

// 	if elem.Type().Kind() != reflect.Struct {
// 		elem = elem.Elem()
// 	}

// 	return elem
// }

// func getJSONFieldName(value *reflect.Value, index int) string {
// 	field := value.Type().Field(index)
// 	json_field := field.Tag.Get("json")
// 	if json_field == "" {
// 		json_field = field.Name
// 	}
// 	json_field = strings.ToLower(json_field)
// 	return json_field
// }

// func makeFieldMap(fields []string) map[string]bool {
// 	fieldMap := make(map[string]bool, len(fields))
// 	for _, s := range fields {
// 		key := strings.ToLower(s)
// 		fieldMap[key] = true
// 	}
// 	return fieldMap
// }

// func fetchResponse(obj interface{}, filterMap map[string]bool) map[string]interface{} {
// 	elem := getElemObject(obj)
// 	fields := map[string]interface{}{}
// 	for i := 0; i < elem.NumField(); i++ {
// 		json_field := getJSONFieldName(&elem, i)
// 		if filterMap == nil || filterMap[json_field] {
// 			fields[json_field] = getElem(elem.Field(i))
// 		}
// 	}
// 	return fields
// }

// func getElem(obj reflect.Value) interface{} {
// 	var response interface{}
// 	switch obj.Type().Kind() {
// 	case reflect.Slice:
// 		slic := []interface{}{}
// 		for i := 0; i < obj.Len(); i++ {
// 			slic = append(slic, getElem(obj.Index(i)))
// 		}
// 		response = slic
// 	case reflect.Interface, reflect.Ptr:
// 		obj.Elem().Interface()
// 	default:
// 		response = obj.Interface()
// 	}
// 	return response
// }

// func validateResponse(response interface{}, expectedResponse interface{}) error {
// 	mapExpectedResponse := fetchResponse(expectedResponse, nil)
// 	fmt.Println(mapExpectedResponse)
// 	fmt.Println(reflect.ValueOf(reflect.TypeOf(mapExpectedResponse["domains"])))
// 	expectedFields := fetchFields(expectedResponse)
// 	expectedFieldMap := makeFieldMap(expectedFields)
// 	fmt.Println(reflect.ValueOf(reflect.TypeOf(response)))
// 	sliceResponse := fetchResponse(response, expectedFieldMap)
// 	fmt.Println(sliceResponse)
// 	fmt.Println(reflect.ValueOf(reflect.TypeOf(sliceResponse["domains"])))
// 	fmt.Println(cmp.Diff(sliceResponse, mapExpectedResponse))
// 	var r DiffReporter
// 	if !cmp.Equal(sliceResponse["domains"], mapExpectedResponse["domains"], cmpopts.EquateEmpty(), cmp.Reporter(&r)) {
// 		return errors.New(r.String())
// 	}

// 	return nil
// }

// // DiffReporter is a simple custom reporter that only records differences
// // detected during comparison.
// type DiffReporter struct {
// 	path  cmp.Path
// 	diffs []string
// }

// func (r *DiffReporter) PushStep(ps cmp.PathStep) {
// 	r.path = append(r.path, ps)
// }

// func (r *DiffReporter) Report(rs cmp.Result) {
// 	if !rs.Equal() {
// 		vx, vy := r.path.Last().Values()
// 		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path.Last().String(), vx, vy))
// 	}
// }

// func (r *DiffReporter) PopStep() {
// 	r.path = r.path[:len(r.path)-1]
// }

// func (r *DiffReporter) String() string {
// 	return strings.Join(r.diffs, "\n")
// }
