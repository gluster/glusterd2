package volgen

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateNode() Node {
	nodes := make([]*Node, 0)
	nodes = append(nodes, &Node{
		ID:      "testing",
		Voltype: "cluster/replicate",
	})
	var n = Node{
		ID:       "testwrite",
		Voltype:  "cluster/replicate",
		Children: nodes,
	}
	return n
}
func TestWrite(t *testing.T) {
	n := generateNode()
	var b bytes.Buffer

	err := n.write(&b)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
}

func TestGraphWrite(t *testing.T) {
	n := generateNode()
	g := Graph{
		id:   "testing",
		root: &n,
	}
	var b bytes.Buffer

	err := g.Write(&b)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
}

func TestWriteToFile(t *testing.T) {
	n := generateNode()
	g := Graph{
		id:   "testing",
		root: &n,
	}

	err := g.WriteToFile("tmp/gd2testing/node.graph")
	assert.NotNil(t, err)

	path := "./test_node.graph"
	defer os.Remove(path)
	err = g.WriteToFile(path)
	assert.Nil(t, err)
}

func TestWriteDot(t *testing.T) {
	n := generateNode()
	var b bytes.Buffer

	err := n.writeDot(&b)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
}

func TestGraphWriteDot(t *testing.T) {
	n := generateNode()
	g := Graph{
		id:   "testing",
		root: &n,
	}
	var b bytes.Buffer

	err := g.WriteDot(&b)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
}

func TestWriteDotToFile(t *testing.T) {
	n := generateNode()
	g := Graph{
		id:   "testing",
		root: &n,
	}

	err := g.WriteDotToFile("tmp/gd2testing/node.graph")
	assert.NotNil(t, err)

	path := "./test_node.graph"
	defer os.Remove(path)
	err = g.WriteDotToFile(path)
	assert.Nil(t, err)
}
