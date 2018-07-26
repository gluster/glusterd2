package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
)

// TODO: Fix this template by making a sub-template and removing markdown
// links if the strings are empty.

const endpointsTable = `
<!---
This file is generated using commands described below. DO NOT EDIT.

$ curl -o endpoints.json -s -X GET http://127.0.0.1:24007/endpoints
$ go build pkg/tools/generate-doc.go
$ ./generate-doc
-->

# REST API Endpoints Reference

**Note:** Fields in request structs marked with "omitempty" struct tag are optional.

Name | Methods | Path | Request | Response
--- | --- | --- | --- | ---{{range $index, $element := .}}
{{.Name}} | {{.Method}} | {{.Path}} | [{{.RequestType}}]({{.BaseURL}}#{{.RequestType}}) | [{{.ResponseType}}]({{.BaseURL}}#{{.ResponseType}}){{end}}
`

// DocEndpoint is a structure the fields of which will be used in the
// template described above. DocEndpoint embeds api.Endpoint and
// extends it by adding a BaseURL that points to the godoc.org site
// with relevant package name as suffix.
type DocEndpoint struct {
	*api.Endpoint
	BaseURL string
}

var pluginMap = map[string]string{
	"Bitro": "plugins/bitrot/api",
	"Devic": "plugins/device/api",
	"Event": "plugins/events/api",
	"GeoRe": "plugins/georeplication/api",
	"SelfH": "plugins/glustershd/api", // TODO: change package name to selfheal
	"Quota": "plugins/quota/api",
	"Rebal": "plugins/rebalance/api",
}

const basePath = "https://godoc.org/github.com/gluster/glusterd2/"

func getGodocURL(endpoint *api.Endpoint) string {

	url := basePath + "pkg/api"

	if pluginPkg, ok := pluginMap[endpoint.Name[:5]]; ok {
		url = basePath + pluginPkg
	}

	return url
}

func generateDocEndpoints(endpoints []api.Endpoint) []DocEndpoint {

	docEndpoints := make([]DocEndpoint, len(endpoints))

	var tmp []string
	var baseURL string
	for i := range endpoints {
		tmp = strings.Split(endpoints[i].RequestType, ".")
		endpoints[i].RequestType = tmp[len(tmp)-1]
		tmp = strings.Split(endpoints[i].ResponseType, ".")
		endpoints[i].ResponseType = tmp[len(tmp)-1]
		baseURL = getGodocURL(&endpoints[i])
		docEndpoints[i] = DocEndpoint{&endpoints[i], baseURL}
	}

	return docEndpoints
}

// TODO: Consider making this code comment instead of markdown in the
// file pkg/api/doc.go to be rendered by godoc in HTML
const outFile = "doc/endpoints.md"

// Example:
// 	$ curl -o endpoints.json -s -X GET http://127.0.0.1:24007/endpoints
//	$ go build pkg/tools/generate-doc.go
//	$ ./generate-doc

func main() {
	var endpointsFile string
	flag.StringVar(&endpointsFile, "endpoints-file", "endpoints.json",
		"The JSON file containing list of endpoints.")
	flag.Parse()

	content, err := ioutil.ReadFile(endpointsFile)
	if err != nil {
		log.Fatal(err)
	}

	var endpoints []api.Endpoint
	if err := json.Unmarshal(content, &endpoints); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	docEndpoints := generateDocEndpoints(endpoints)

	t := template.Must(template.New("endpoints").Parse(endpointsTable))
	if err := t.Execute(f, docEndpoints); err != nil {
		log.Fatal(err)
	}
}
