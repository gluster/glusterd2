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

	return json.Unmarshal(body, v)
}

// GetJSONFromRequest unmarshals JSON in `r` into `v`
func GetJSONFromRequest(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return jsonFromBody(r.Body, v)
}

// GetJSONFromResponse unmarshals JSON in `r` into `v`
func GetJSONFromResponse(r *http.Response, v interface{}) error {
	defer r.Body.Close()
	return jsonFromBody(r.Body, v)
}
