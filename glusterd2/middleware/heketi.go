package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/pkg/utils"
)

// TODO
// In Go, the idiomatic and recommended way to attach any request scoped
// metadata information across goroutine and process boundaries is to use the
// 'context' package. This is not useful unless we pass down this context
// all through-out the request scope across nodes, and this involves some
// code changes in function signatures at many places
// The following simple implementation is good enough until then...

// Heketi is a middleware which generates adds bricks to a volume
// request if it has a key asking for auto brick allocation. It modifies the
// HTTP request and adds bricks to it.
func Heketi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {


		if r.URL.Path == "/v1/volumes" && r.Method == http.MethodPost {
			req := new(volumecommands.VolCreateRequest)
			utils.GetJSONFromRequest(r, req)

			//if (len(req.Bricks) <= 0) && (req.Size > 0) {
			if req.Size > 0 {
				replacer := strings.NewReplacer("export", "testexport")
				req.Bricks[0] = replacer.Replace(req.Bricks[0])
				req.Bricks[1] = replacer.Replace(req.Bricks[1])
				req.Bricks[2] = replacer.Replace(req.Bricks[2])
				req.Bricks[3] = replacer.Replace(req.Bricks[3])
				req.Bricks[4] = replacer.Replace(req.Bricks[4])
				req.Bricks[5] = replacer.Replace(req.Bricks[5])

				newbody, err := json.Marshal(req)
				if err != nil {
					fmt.Printf("Marshalling Error %v", err)
				}

				r.Body = ioutil.NopCloser(bytes.NewReader(newbody))
				r.ContentLength = int64(len(newbody))

			}

		}
		next.ServeHTTP(w, r)
	})
}
