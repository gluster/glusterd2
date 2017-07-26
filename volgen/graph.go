package volgen

type Node struct {
	Voltype  string
	Id       string
	Children []*Node
	Options  map[string]string
}

type Graph struct {
	id   string
	root *Node
}

func NewGraph() *Graph {
	return new(Graph)
}

func NewNode() *Node {
	n := new(Node)
	n.Options = make(map[string]string)
	return n
}
