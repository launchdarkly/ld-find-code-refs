package element

type ElementMatcher struct {
	Elements   []string
	Aliases    map[string][]string
	Delimiters []string
}

type ElementsMatcher struct {
	Elements    []ElementMatcher
	Type        string
	CtxLines    int
	Delimiters  string
	Directories []string
}
