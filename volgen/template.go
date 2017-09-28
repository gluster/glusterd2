package volgen

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
)

const (
	templateExt = ".graph"
)

// Templates are empty graphs built from template files
type GraphTemplate Graph

type TemplateNotFoundError string

var templates map[string]*GraphTemplate

// LoadTemplates reads and loads all the templates from the template directory
func LoadTemplates() error {
	tdir := config.GetString(templateDirOpt)
	log.WithField("templatesdir", tdir).Debug("loading templates")
	glob := path.Join(tdir, "*"+templateExt)

	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	log.WithField("templates", fs).Debug("found templates")

	templates = make(map[string]*GraphTemplate)
	for _, f := range fs {
		gt, err := ReadTemplateFile(f)
		if err != nil {
			return err
		}
		templates[path.Base(f)] = gt
		log.WithField("template", f).Debug("loaded template")
	}

	return nil
}

// ReadTemplateFile reads in a template file and generates a template graph
func ReadTemplateFile(p string) (*GraphTemplate, error) {
	tf, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer tf.Close()

	t := new(GraphTemplate)
	t.id = path.Base(p)

	var curr, prev *Node

	s := bufio.NewScanner(tf)
	for s.Scan() {
		curr = NewNode()

		curr.Voltype = s.Text()
		curr.ID = path.Base(curr.Voltype)
		if t.root == nil {
			t.root = curr
		}
		if prev != nil {
			prev.Children = append(prev.Children, curr)
		}
		prev = curr
		// TODO: Handle graph templates with branches
	}

	return t, nil
}

func getTemplate(n string) (*GraphTemplate, error) {
	t, ok := templates[n]
	if !ok {
		return nil, TemplateNotFoundError(n)
	}
	return t, nil
}

// Error returns the error string for TemplateNotFoundError
func (t TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", string(t))
}
