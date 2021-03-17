package hcl2template

// ComponentKind helps enumerate what kind of components exist in this Package.
type ComponentKind int

const (
	Builder ComponentKind = iota
	Provisioner
	PostProcessor
	Datasource
)
