package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func jsonFromBody(r io.Reader, v interface{}) error {

	// Check body
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}

func GetJsonFromRequest(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return jsonFromBody(r.Body, v)
}

func GetJsonFromResponse(r *http.Response, v interface{}) error {
	defer r.Body.Close()
	return jsonFromBody(r.Body, v)
}
