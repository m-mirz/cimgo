package main

// fileSpec is the data passed to validation_file.tmpl.
type fileSpec struct {
	FileName         string
	Pkg              string
	OrchestratorName string
	Imports          []string // sorted, deduplicated; always includes cimgo/cimstructs
	Checks           []checkSpec
}

// checkSpec describes a single Check<...> function. Guard and Condition carry
// the only varying pieces of the function body; everything else is fixed
// scaffolding emitted by the template. Decl is an optional package-level
// declaration emitted directly above the function (used by Pattern checks to
// hoist the compiled regexp out of the per-call hot path). Prelude is an
// optional block emitted before the main per-element loop — used by inverse
// path checks to build an O(N) cross-reference index once per Check, instead
// of paying O(N²) by scanning inside the loop. NoV switches the type assertion
// from `v, ok := ...` to `_, ok := ...` for checks that don't need the
// instance value (typical for inverse-path checks that only consume `id`).
type checkSpec struct {
	Name         string
	ShapeID      string // Original SHACL Shape ID (e.g. eqc:ACLineSegment.length-length)
	RuleID       string // Extracted Rule ID (e.g. eq600:ACLineSegment.length-length)
	RuleName     string
	Description  string
	Class        string
	Tag          string
	Component    string
	Property     string
	Message      string
	Severity     string
	Decl         string // optional package-level declaration emitted before the function
	Prelude      string // optional block emitted before the main loop (e.g. inverse-ref index)
	NoV          bool   // suppress the v binding when the loop body doesn't use it
	Guard         string // tab-indented, may span multiple lines; empty if none
	Condition     string // single expression; opens the violation block as `if <Condition> {`
	DatasetCheck  bool   // emit a single dataset-level check (no per-element loop) when true
	SelfContained bool   // Guard appends violations directly; skip the outer `if Condition` block
}
