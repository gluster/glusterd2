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

// GraphTemplate are empty graphs built from template files
type GraphTemplate Graph

// TemplateNotFoundError is returned by GetTemplate when the specified template
// cannot be found
type TemplateNotFoundError string

var (
	templates            map[string]*GraphTemplate
	defaultTemplatePaths map[string]string
)

func init() {
	templates = make(map[string]*GraphTemplate)
	defaultTemplatePaths = make(map[string]string)
}

// LoadTemplates reads and loads all the templates from the default template directory
// and sets up the default graph map
func LoadTemplates() error {
	tdir := config.GetString(templateDirOpt)
	log.WithField("templatesdir", tdir).Debug("loading templates")
	glob := path.Join(tdir, "*"+templateExt)

	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	log.WithField("templates", fs).Debug("found templates")

	for _, f := range fs {
		_, err := LoadTemplate(f)
		if err != nil {
			return err
		}
		defaultTemplatePaths[path.Base(f)] = f
	}

	return nil
}

// LoadTemplate loads the template at the given path
func LoadTemplate(path string) (*GraphTemplate, error) {
	gt, err := ReadTemplateFile(path)
	if err != nil {
		return nil, err
	}
	templates[path] = gt
	log.WithField("template", path).Debug("loaded template")

	return gt, nil
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

// GetTemplate returns the specified graph template.
func GetTemplate(id string, umap map[string]string) (*GraphTemplate, error) {
	var (
		path string
		ok   bool
		t    *GraphTemplate
		err  error
	)

	// Find the template path in the usermap and the defaultGraphMap in order
	path, ok = umap[id]
	if !ok {
		path, ok = defaultTemplatePaths[id]
		if !ok {
			return nil, TemplateNotFoundError(id)
		}
	}

	// Get template from templates map, if not found load user template
	t, ok = templates[path]
	if !ok {
		// TODO: Ensure that user template is in safe paths
		t, err = LoadTemplate(path)
		if err != nil {
			return nil, TemplateNotFoundError(id)
		}
	}
	return t, nil
}

// Error returns the error string for TemplateNotFoundError
func (t TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", string(t))
}
