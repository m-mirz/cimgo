package sparql

import (
	"cimgo/rdf/graph"
	"cimgo/rdf/term"
)

// Result holds the result of a SPARQL query.
// Ported from: rdflib.plugins.sparql.sparql.Query result types
type Result struct {
	Type      string                 // "SELECT", "ASK", "CONSTRUCT"
	Vars      []string               // variable names for SELECT
	Bindings  []map[string]term.Term // solution mappings for SELECT
	AskResult bool                   // result for ASK
	Graph     *graph.Graph           // result graph for CONSTRUCT
}
