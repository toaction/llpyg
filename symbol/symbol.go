package symbol

type Symbol struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Doc  string    `json:"doc"`
	Sig  Signature `json:"sig"`
}

type Module struct {
	Name      string    `json:"name"`      // python module name
	Functions []*Symbol `json:"functions"` // package functions
	// TODO: variables, classes, etc.
}

type SigSource int

const (
	SigSourceDoc SigSource = iota
	SigSourceInspect
	SigSourceParadigm
)

type Signature struct {
	Str    string    `json:"str"`
	Source SigSource `json:"source"`
}
