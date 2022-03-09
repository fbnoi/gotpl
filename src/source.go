package template

func NewSource(code, name, path string) *Source {
	return &Source{
		Code: code,
		Name: name,
		Path: path,
	}
}

type Source struct {
	Code string
	Name string
	Path string
}
