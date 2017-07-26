package volgen

type ErrOptsNotFound string
type ErrOptRequired string

func (e ErrOptsNotFound) Error() string {
	return "options not found for given xlator: " + string(e)
}

func (e ErrOptRequired) Error() string {
	return "option '" + string(e) + "' is required and needs to be set explicitly"
}
