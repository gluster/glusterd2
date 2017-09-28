package volgen

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

const (
	volfileTemplate = `{{define "volume" -}}
volume {{.ID}}
	type {{.Voltype}}
	{{- range $key, $val := .Options}}
	option {{$key}} {{$val}}
	{{- else}}
	{{- end}}
	{{- if .Children}}
	subvolumes{{range $child := .Children}} {{$child.ID}}{{end}}
	{{- else}}
	{{- end}}
end-volume
{{end}}
`
	dotfileTemplate = `{{- define "volume"}}
{{- $node := . }}
{{- range $child := .Children}}
"{{$node.ID}}" -> "{{$child.ID}}"
{{- end}}
{{end}}
`
	dotHeader  = "digraph {"
	dotTrailer = "\n}"
)

var (
	volTmpl = template.Must(template.New("volume").Parse(volfileTemplate))
	dotTmpl = template.Must(template.New("volume").Parse(dotfileTemplate))
)

// Write will write the graph to the given writer
func (n *Node) write(w io.Writer) error {
	for _, c := range n.Children {
		c.write(w)
		fmt.Fprintln(w)
	}
	return volTmpl.Execute(w, n)
}

// Write write the volfile to the given io.Writer
func (g *Graph) Write(w io.Writer) error {
	return g.root.write(w)
}

// WriteToFile writes the graph to the given path, creating the volfile.
// NOTE: Any existing file at the path is truncated.
func (g *Graph) WriteToFile(path string) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if e != nil {
		return e
	}
	defer f.Close()
	defer f.Sync()

	return g.Write(f)
}

func (n *Node) writeDot(w io.Writer) error {
	for _, c := range n.Children {
		c.writeDot(w)
	}
	return dotTmpl.Execute(w, n)
}

// WriteDot writes a dot graph of volume to the writer
func (g *Graph) WriteDot(w io.Writer) error {
	w.Write([]byte(dotHeader))
	g.root.writeDot(w)
	w.Write([]byte(dotTrailer))

	return nil
}

// WriteDotToFile writes the dot graph to the given path.
// NOTE: Any existing file at the path is truncated
func (g *Graph) WriteDotToFile(path string) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if e != nil {
		return e
	}
	defer f.Close()
	defer f.Sync()

	return g.WriteDot(f)
}
