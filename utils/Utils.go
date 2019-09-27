package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
)

var Env = "development"

type JSONErrror struct {
	Msg string `json:"msg"`
}

func WriteJSONError(w http.ResponseWriter, Msg string) {
	log.Printf("%v", Msg)
	w.WriteHeader(http.StatusBadRequest)
	jsonError := JSONErrror{Msg}
	json.NewEncoder(w).Encode(jsonError)
}

func SelectFields(s interface{}, fields ...string) map[string]interface{} {
	fs := fieldSet(fields...)
	rt, rv := reflect.TypeOf(s), reflect.ValueOf(s)
	out := make(map[string]interface{}, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonKey := field.Tag.Get("json")
		if fs[jsonKey] {
			out[jsonKey] = rv.Field(i).Interface()
		}
	}
	return out
}

func fieldSet(fields ...string) map[string]bool {
	set := make(map[string]bool, len(fields))
	for _, s := range fields {
		set[s] = true
	}
	return set
}

func DecodeMsg(r *http.Request, object interface{}, disallow bool) (interface{}, error) {
	if r.Body == nil {
		return nil, errors.New("no request body")
	}

	decoder := json.NewDecoder(r.Body)
	if disallow {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func SetEnv(newEnvironment string) {
	Env = newEnvironment
}

// HealthCheck godoc
// @Summary Health check
// @Description health check
// @Tags health
// @Success 200
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
