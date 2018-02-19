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

Name | Methods | Path | Request | Response
--- | --- | --- | --- | ---{{range $index, $element := .}}
{{.Name}} | {{.Method}} | {{.Path}} | [{{.RequestType}}](https://godoc.org/github.com/gluster/glusterd2/pkg/api#{{.RequestType}}) | [{{.ResponseType}}](https://godoc.org/github.com/gluster/glusterd2/pkg/api#{{.ResponseType}}){{end}}
`

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

	var tmp []string
	for i := range endpoints {
		tmp = strings.Split(endpoints[i].RequestType, ".")
		endpoints[i].RequestType = tmp[len(tmp)-1]
		tmp = strings.Split(endpoints[i].ResponseType, ".")
		endpoints[i].ResponseType = tmp[len(tmp)-1]
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	t := template.Must(template.New("endpoints").Parse(endpointsTable))
	if err := t.Execute(f, endpoints); err != nil {
		log.Fatal(err)
	}
}
