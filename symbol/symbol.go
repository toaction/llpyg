package symbol

// represents the origin of a function signature
type SigSource int

const (
	// signature extracted from docstring
	SigSourceDoc SigSource = iota
	// signature from Python inspect module
	SigSourceInspect
	// generic fallback signature
	SigSourceParadigm
)

// represents a function signature with its source information
type Signature struct {
	Str    string    `json:"str"`    // The signature string
	Source SigSource `json:"source"` // Where the signature was obtained
}

type Symbol struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Doc  string    `json:"doc"`
	Sig  Signature `json:"sig"`
}

type Module struct {
	Name      string    `json:"name"`      // python module name
	Functions []*Symbol `json:"functions"` // package functions
	Classes   []*Class  `json:"classes"`
}

// base class
type Base struct {
	Name   string `json:"name"`
	Module string `json:"module"`
}

// @property
type Property struct {
	Name   string    `json:"name"`
	Getter Signature `json:"getter"`
	Setter Signature `json:"setter"`
}

// Python class
type Class struct {
	Name            string      `json:"name"`
	Doc             string      `json:"doc"`
	Bases           []*Base     `json:"base"`
	InitMethod      *Symbol     `json:"initMethod"`
	InstanceMethods []*Symbol   `json:"instanceMethods"` // include override special methods
	ClassMethods    []*Symbol   `json:"classMethods"`
	StaticMethods   []*Symbol   `json:"staticMethods"`
	Properties      []*Property `json:"properties"`
	// TODO: attributes
}
