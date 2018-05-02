package volgen

import (
	"io/ioutil"
	"os"
	"testing"

	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadTemplates(t *testing.T) {
	tmpPath := "/tmp/gd2temptest"
	config.Set("templatesdir", tmpPath)

	err := LoadTemplates()
	assert.Nil(t, err)
	defer os.RemoveAll(tmpPath)

	files, _ := ioutil.ReadDir(tmpPath)
	assert.NotEmpty(t, len(files))

}

func TestLoadTemplate(t *testing.T) {
	gt, err := LoadTemplate("/tmp/testing")
	assert.NotNil(t, err)
	assert.Nil(t, gt)

	gt, err = LoadTemplate("./templates/brick.graph")
	assert.Nil(t, err)
	assert.NotNil(t, gt)
}

func TestReadTemplateFile(t *testing.T) {
	gt, err := ReadTemplateFile("/tmp/testing")
	assert.Nil(t, gt)
	assert.Contains(t, err.Error(), "no such file")

	gt, err = ReadTemplateFile("./templates/brick.graph")
	assert.Nil(t, err)
	assert.NotNil(t, gt)
}

func TestGetTemplate(t *testing.T) {
	umap := make(map[string]string)

	gt, err := GetTemplate("testID", umap)
	assert.Nil(t, gt)
	assert.Contains(t, err.Error(), "template not found")

	tmpPath := "/tmp/gd2temptest"
	config.Set("templatesdir", tmpPath)

	err = LoadTemplates()
	assert.Nil(t, err)

	defer os.RemoveAll(tmpPath)
	umap["brick.graph"] = "/tmp/gd2temptest/brick.graph"
	gt, err = GetTemplate("brick.graph", umap)
	assert.Nil(t, err)
	assert.NotNil(t, gt)

}

func TestTemplateNotFound(t *testing.T) {
	var tmp TemplateNotFoundError = "test"
	s := tmp.Error()
	assert.Contains(t, s, "template not found")
}
