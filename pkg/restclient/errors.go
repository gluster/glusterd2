package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// HTTPErrorResponse is custom error that is returned when expected
// status code does not match with actual status code returned.
type HTTPErrorResponse struct {
	Status  int
	Body    string
	Headers http.Header
}

func (e *HTTPErrorResponse) Error() string {
	return fmt.Sprintf("Request failed. Status: %d\nResponse: %s", e.Status, e.Body)
}

func newHTTPErrorResponse(resp *http.Response) error {

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var errResp api.ErrorResp
	if err = json.Unmarshal(b, &errResp); err != nil {
		return err
	}

	var buffer bytes.Buffer
	// FIXME: The CLI should be doing this string processing.
	for _, apiErr := range errResp.Errors {
		switch api.ErrorCode(apiErr.Code) {
		case api.ErrTxnStepFailed:
			buffer.WriteString(fmt.Sprintf(
				"Transaction step %s failed on peer %s with error: %s\n",
				apiErr.Fields["step"], apiErr.Fields["peer-id"], apiErr.Fields["error"]))
		default:
			buffer.WriteString(apiErr.Message)
		}
	}

	return fmt.Errorf("%s", buffer.String())
}
