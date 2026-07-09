// Static call-graph analysis over the hand-written SPARQL validation package
// (default: `validation/`), collecting every distinct sh:name reachable from
// each profile group's entry point(s). Used by `cmd/shaclgen -rule-report`,
// combined in ttl_report.go with the SPARQL constraint shapes actually
// defined in the CGMES SHACL TTL files, to regenerate README.md's "SPARQL
// Check Coverage" table instead of hand-maintaining it. Ported from
// cimoxide's cimgen/src/shacl/sparql_report.rs, adapted to Go: since every
// hand-written check lives in a single package, function names are unique
// package-wide, so call resolution needs no per-file/module qualification.
//
// Matching is done on the Violation.Name field (the CGMES conformance rule
// name, e.g. "C:452:EQ:SynchronousMachine:aggregate") rather than RuleID: the
// TTL side's sh:name is a plain string with no namespace to normalize, unlike
// the SHACL shape IRI backing RuleID, whose prefix can legitimately differ
// between the TTL file's own declaration and however the importer
// canonicalizes it.
//
// A plain scan for `Violation{Name: "...", ...}` literals is not enough:
// some check functions build the literal directly, but others (e.g.
// sparql_dynamics.go's CheckGovHydro4GainPoints) delegate to a local closure
// constructor that takes the name as a parameter and is called once per
// constraint with a literal argument -- the Go analogue of cimoxide's
// prof10_violation constructor -- or assign it to a local variable across
// branches before building the Violation (e.g.
// sparql_ssh_notsolvedmas.go's checkCsConverterTargetAngleApplicability).
// Both patterns are detected so the literal is read back out from the call
// site or assignment instead of the (non-literal) field where it's used.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// group lists the entry-point function names whose reachable check functions
// make up one row of the README "SPARQL Check Coverage" table. Matches GROUPS
// in cimgen/src/shacl/sparql_report.rs row-for-row: Common/AllProfiles absorbs
// C:600 conformance (both cross-cutting, not tied to one profile), and
// Topology/DiagramLayout/Operation each get their own row.
type sparqlGroup struct {
	Label   string
	Entries []string
}

var sparqlGroups = []sparqlGroup{
	{"Equipment (EQ)", []string{
		"ValidateEQProfileSPARQL",
		"ValidateEQNotSolvedMASProfileSPARQL",
		"ValidateEQBDProfileSPARQL",
	}},
	{"Steady State Hypothesis (SSH)", []string{
		"ValidateSSHProfileSPARQL",
		"ValidateSSHNotSolvedMASProfileSPARQL",
	}},
	{"Dynamics (DY)", []string{"ValidateDYProfileSPARQL"}},
	{"State Variables (SV)", []string{
		"ValidateSVProfileSPARQL",
		"ValidateSVSolvedMASProfileSPARQL",
	}},
	{"Short Circuit (SC)", []string{
		"ValidateSCProfileSPARQL",
		"ValidateSCNotSolvedMASProfileSPARQL",
	}},
	// C:600 conformance (ValidateProf10HeaderRules) has no row of its own: it's a
	// cross-cutting rule like Common/AllProfiles, not tied to one profile, and it's
	// already reached transitively from ValidateCommonRulesSPARQL's call graph, so no
	// extra wiring is needed to fold it into this group.
	{"Common / AllProfiles", []string{
		"ValidateCommonRulesSPARQL", // transitively reaches ValidateProf10HeaderRules
		"ValidateCommonRulesSolvedMASSPARQL",
	}},
	{"Topology (TP)", []string{"ValidateTPNotSolvedMASProfileSPARQL"}},
	{"DiagramLayout (DL)", []string{"ValidateDLProfileSPARQL"}},
	{"Operation (OP)", []string{"ValidateOPProfileSPARQL"}},
	// Not part of the historical SPARQL Check Coverage table, but reported the
	// same way since these checks carry real rule names too.
	// CheckBaseVoltageInEQBD is invoked directly from all_rules.go's
	// RunValidation, not through ValidateCIMdeskQualityChecks, so it needs its
	// own entry point.
	{"CIMdesk quality", []string{"ValidateCIMdeskQualityChecks", "CheckBaseVoltageInEQBD"}},
}

// sparqlGroupReport is one row of the coverage report.
type sparqlGroupReport struct {
	Label string
	// Names is every distinct Violation.Name string reachable from the
	// group's entry points.
	Names []string
}

// sparqlReport walks every *.go file in dir (default "validation") and
// computes one sparqlGroupReport per entry in sparqlGroups.
func sparqlReport(dir string) ([]sparqlGroupReport, error) {
	fns, err := buildFnIndex(dir)
	if err != nil {
		return nil, err
	}
	constructors := detectConstructors(fns)

	var out []sparqlGroupReport
	for _, g := range sparqlGroups {
		visited := map[string]bool{}
		nameSet := map[string]bool{}
		for _, entry := range g.Entries {
			collectNames(fns, constructors, entry, visited, nameSet)
		}
		names := make([]string, 0, len(nameSet))
		for n := range nameSet {
			names = append(names, n)
		}
		sort.Strings(names)
		out = append(out, sparqlGroupReport{Label: g.Label, Names: names})
	}
	return out, nil
}

// buildFnIndex parses every non-test *.go file directly under dir and
// indexes every top-level function declaration by name. Function names are
// unique package-wide, so a flat map (unlike cimoxide's per-file keying) is
// sufficient.
func buildFnIndex(dir string) (map[string]*ast.FuncDecl, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read validation directory %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	fset := token.NewFileSet()
	fns := map[string]*ast.FuncDecl{}
	for _, name := range names {
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || fd.Body == nil {
				continue
			}
			fns[fd.Name.Name] = fd
		}
	}
	return fns, nil
}

// detectConstructors finds package-level functions whose body directly
// returns/builds a `Violation{Name: <param>, ...}` literal -- i.e. the field
// is populated straight from a parameter, so the literal has to be read from
// each call site instead. detection is kept general and symmetric with
// localClosureConstructors in case one is added later.
func detectConstructors(fns map[string]*ast.FuncDecl) map[string]int {
	out := map[string]int{}
	for name, fn := range fns {
		if fn.Type == nil {
			continue
		}
		if idx, ok := constructorParamIndex(fn.Type.Params, fn.Body); ok {
			out[name] = idx
		}
	}
	return out
}

// localClosureConstructors finds function-literal closures assigned to a
// local variable inside fn (`name := func(...) {...}`) that follow the same
// constructor shape, e.g. sparql_dynamics.go's CheckGovHydro4GainPoints
// defines `checkZero := func(val float64, prop, ruleID, name string) {...}`
// and calls it once per constraint with a literal name argument. The
// returned map is only valid for call sites within fn, since the closure is
// not visible outside it.
func localClosureConstructors(fn *ast.FuncDecl) map[string]int {
	out := map[string]int{}
	if fn.Body == nil {
		return out
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.DEFINE {
			return true
		}
		for i, rhs := range assign.Rhs {
			fl, ok := rhs.(*ast.FuncLit)
			if !ok || i >= len(assign.Lhs) {
				continue
			}
			name, ok := assign.Lhs[i].(*ast.Ident)
			if !ok {
				continue
			}
			if idx, ok := constructorParamIndex(fl.Type.Params, fl.Body); ok {
				out[name.Name] = idx
			}
		}
		return true
	})
	return out
}

// constructorParamIndex reports whether body's first Violation{...} literal
// assigns its Name field from one of params, and if so, which parameter
// position. Shared by detectConstructors (package-level functions) and
// localClosureConstructors (local closures).
func constructorParamIndex(params *ast.FieldList, body *ast.BlockStmt) (int, bool) {
	if params == nil || body == nil {
		return 0, false
	}
	var paramNames []string
	for _, field := range params.List {
		if len(field.Names) == 0 {
			paramNames = append(paramNames, "")
			continue
		}
		for _, n := range field.Names {
			paramNames = append(paramNames, n.Name)
		}
	}

	var lit *ast.CompositeLit
	ast.Inspect(body, func(n ast.Node) bool {
		if lit != nil {
			return false
		}
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if id, ok := cl.Type.(*ast.Ident); ok && id.Name == "Violation" {
			lit = cl
			return false
		}
		return true
	})
	if lit == nil {
		return 0, false
	}

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Name" {
			continue
		}
		id, ok := kv.Value.(*ast.Ident)
		if !ok {
			continue
		}
		for i, p := range paramNames {
			if p == id.Name {
				return i, true
			}
		}
	}
	return 0, false
}

// literalStringArg unwraps a parenthesized string literal expression down to
// its string value, if any.
func literalStringArg(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		if err != nil {
			return "", false
		}
		return s, true
	case *ast.ParenExpr:
		return literalStringArg(v.X)
	default:
		return "", false
	}
}

// directTargets returns the distinct, resolvable call targets found anywhere
// in fn's own body (does not recurse into the callees' own bodies -- fn.Body
// is only this function's syntax tree, so there is nothing else to
// accidentally walk into). Constructor calls are excluded -- they contribute
// a name at the call site (see violationNames), not a further call target.
// Used by collectNames to walk the call graph reachable from each group's
// entry point(s).
func directTargets(fn *ast.FuncDecl, fns map[string]*ast.FuncDecl, constructors map[string]int) []string {
	seen := map[string]bool{}
	var out []string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		if _, isConstructor := constructors[id.Name]; isConstructor {
			return true
		}
		if _, known := fns[id.Name]; known && id.Name != fn.Name.Name && !seen[id.Name] {
			seen[id.Name] = true
			out = append(out, id.Name)
		}
		return true
	})
	return out
}

// localVarLiterals collects every string literal directly assigned (`=`) to
// a local variable anywhere in fn's body. Used as a fallback when a
// Violation{Name: x} field's value is a plain local variable rather than a
// literal or a known constructor call -- e.g.
// sparql_ssh_notsolvedmas.go's checkCsConverterTargetAngleApplicability
// branches into `ruleName = "C:301:SSH:CsConverter.targetAlpha:applicability"`
// or `ruleName = "C:301:SSH:CsConverter.targetGamma:applicability"` depending
// on a parameter before building the Violation, so both literals need to be
// attributed to any use of `ruleName` in a Name field, not just one.
func localVarLiterals(fn *ast.FuncDecl) map[string][]string {
	out := map[string][]string{}
	if fn.Body == nil {
		return out
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.ASSIGN {
			return true
		}
		for i, rhs := range assign.Rhs {
			if i >= len(assign.Lhs) {
				continue
			}
			id, ok := assign.Lhs[i].(*ast.Ident)
			if !ok {
				continue
			}
			if v, ok := literalStringArg(rhs); ok && v != "" {
				out[id.Name] = append(out[id.Name], v)
			}
		}
		return true
	})
	return out
}

// violationNames returns every Violation.Name string reachable directly in
// fn's own body: literals assigned to the Name field of a Violation{...}
// composite literal (directly, or via an intermediate local variable -- see
// localVarLiterals), plus literals passed to any constructor call (a
// package-level constructor, or a closure local to fn -- see
// localClosureConstructors).
//
// A Name literal is split on "|" before being recorded: some hand-written
// checks copy a compound sh:name verbatim (e.g. sparql_common.go's
// CheckIdentifiedObjectStringLengths sets Name to
// "C:301:EQ:IdentifiedObject.shortName:stringLength|C:301:EQBD:...|...",
// covering several profile-specific conformance rules with one Violation),
// matching how ttl_report.go splits the TTL's own compound sh:name into
// individual candidates.
func violationNames(fn *ast.FuncDecl, constructors map[string]int) []string {
	if fn.Body == nil {
		return nil
	}
	local := localClosureConstructors(fn)
	localVars := localVarLiterals(fn)
	lookup := func(name string) (int, bool) {
		if idx, ok := local[name]; ok {
			return idx, true
		}
		idx, ok := constructors[name]
		return idx, ok
	}

	var out []string
	add := func(s string) {
		for part := range strings.SplitSeq(s, "|") {
			if part != "" {
				out = append(out, part)
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			id, ok := node.Type.(*ast.Ident)
			if !ok || id.Name != "Violation" {
				return true
			}
			for _, elt := range node.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "Name" {
					continue
				}
				if v, ok := literalStringArg(kv.Value); ok && v != "" {
					add(v)
					continue
				}
				if vid, ok := kv.Value.(*ast.Ident); ok {
					for _, v := range localVars[vid.Name] {
						add(v)
					}
				}
			}
		case *ast.CallExpr:
			id, ok := node.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			idx, ok := lookup(id.Name)
			if !ok || idx >= len(node.Args) {
				return true
			}
			if v, ok := literalStringArg(node.Args[idx]); ok && v != "" {
				add(v)
			}
		}
		return true
	})
	return out
}

// collectNames recursively visits every function reachable from name and
// accumulates every Violation.Name string it finds.
func collectNames(fns map[string]*ast.FuncDecl, constructors map[string]int, name string, visited map[string]bool, out map[string]bool) {
	if visited[name] {
		return
	}
	visited[name] = true
	fn, ok := fns[name]
	if !ok {
		return
	}
	for _, n := range violationNames(fn, constructors) {
		out[n] = true
	}
	for _, t := range directTargets(fn, fns, constructors) {
		collectNames(fns, constructors, t, visited, out)
	}
}
