package shaclimport

const DefaultSHACLPattern = "application-profiles-library/CGMES/CurrentRelease/SHACL/*.ttl"

// ConstraintInfo is the simplified representation of a single SHACL constraint,
// produced by ProcessFileToResults + SimplifyFileResults and consumed by callers.
type ConstraintInfo struct {
	Path        []string       `json:"path"`
	Severity    string         `json:"severity,omitempty"`
	Message     string         `json:"message,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Component   string         `json:"component"`
	Payload     map[string]any `json:"payload"`
}

func (c ConstraintInfo) IsSPARQL() bool {
	return c.Component == "sh:SPARQLConstraintComponent"
}

func (c ConstraintInfo) IsSHACL() bool {
	return !c.IsSPARQL()
}

type TargetInfo struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type SparqlValuesInfo struct {
	Select   string `json:"select"`
	Prefixes string `json:"prefixes"`
	Expr     string `json:"expr,omitempty"`
}

type ShapeInfo struct {
	ID          string            `json:"id"`
	Targets     []TargetInfo      `json:"targets,omitempty"`
	Path        []string          `json:"path,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Constraints []ConstraintInfo  `json:"constraints,omitempty"`
	Properties  []ShapeInfo       `json:"properties,omitempty"`
	Values      *SparqlValuesInfo `json:"values,omitempty"`
	Severity    string            `json:"severity,omitempty"`
	Messages    []string          `json:"messages,omitempty"`
}

type FileResults struct {
	FileName string      `json:"file_name"`
	Shapes   []ShapeInfo `json:"shapes"`
}

// SimplifiedDrop records a constraint dropped during SimplifyFileResults.
type SimplifiedDrop struct {
	Classes   []string
	Prop      string
	Component string
	Name      string
	Reason    string
}

type FileStats struct {
	Name         string
	ShaclPath    string
	SparqlPath   string
	ShaclCounts  map[string]int
	SparqlCounts map[string]int
}
