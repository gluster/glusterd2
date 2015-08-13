package volgen

type Xlator struct {
	Name     string
	Type     string
	Options  map[string]string
	Children []Xlator
}
