package volgen

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const (
	templateExt = ".graph"
)

// GraphTemplate are empty graphs built from template files, which define the
// basic structure of a GlusterFS volume graph
//
// Template files are simple text files which contain a list of xlators.
// Each xlator should be specified on its own line, and in order from root of the graph to the leaf.
// For example,
// 	protocol/server
// 	performance/decompounder
// 	debug/io-stats
// 	.
// 	.
// 	.
//
// In addition to xlator name, each line can also specify an alternate name to
// be used to name the xlator in generated graphs.  Alternate names are
// specified following the xlator, separated by a comma.
// Alternate names can also use varstrings. If the alternate name is a
// varstring, the xlator will be named as the replacement of the varstring. If
// not xlator will be named "<volname>-<altname>".
// For example,
// 	performance/decompounder, {{ brick.path }}
//
// For now only linear graph strcutures are possible.
// TODO: Improve template to support branches
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

	var fs []string
	fs, _ = filepath.Glob(glob)

	// Generate default templates if not exists
	os.MkdirAll(tdir, os.ModePerm)
	for _, g := range defaultGraphs {
		p := path.Join(tdir, g.name)
		_, err := os.Stat(p)
		if err != nil && os.IsNotExist(err) {
			if err = ioutil.WriteFile(p, []byte(g.content), 0644); err != nil {
				log.WithField("file", p).WithError(err).Error("failed to generate default template")
			}
			fs = append(fs, p)
		}
	}
	log.WithField("templates", fs).Debug("generated default templates")

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

		// Split line into 2 parts
		// 1st part is the xlator and what ever remains is the altname
		// Altnames can be varstrings
		// TODO: Have a better way to do this tokenization than just string splitting
		tokens := strings.SplitN(s.Text(), ",", 2)

		curr.Voltype = tokens[0]
		if len(tokens) == 2 {
			curr.ID = tokens[1]
		} else {
			curr.ID = path.Base(curr.Voltype)
		}

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
