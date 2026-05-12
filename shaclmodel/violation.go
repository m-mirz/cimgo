// Package shaclmodel holds shared types for SHACL validation results. It is a
// leaf package so that both the hand-written validation code and the generated
// shaclgen package can depend on it without forming an import cycle.
package shaclmodel

// Violation describes a single failed SHACL constraint against one focus node.
type Violation struct {
	ObjectID    string
	RuleID      string
	Class       string
	Property    string
	Message     string
	Severity    string
	Name        string
	Description string
}
