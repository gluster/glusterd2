package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
)

const endpointsTable = `# API Endpoints

Name | Methods | Path
--- | --- | ---{{range $index, $element := .}}
{{.Name}} | {{.Methods}} | {{.Path}}{{end}}
`

const outFile = "endpoints.md"

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

	t := template.Must(template.New("endpoints").Parse(endpointsTable))
	if err := t.Execute(f, endpoints); err != nil {
		log.Fatal(err)
	}
}
