package sparql_test

import (
	"fmt"
	"strings"
	"testing"

	"cimgo/rdf/graph"
	"cimgo/rdf/sparql"
	"cimgo/rdf/term"
	"cimgo/rdf/turtle"
)

func TestTripleTermSpecialChars(t *testing.T) {
	g := graph.NewGraph()
	g.Add(
		term.NewURIRefUnsafe("http://ex/s"),
		term.NewURIRefUnsafe("http://ex/p"),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral("hello \"world\" \n\t", term.WithDatatype(term.XSDString)),
		),
	)
	r, err := sparql.Query(g, `SELECT ?o WHERE { <http://ex/s> <http://ex/p> ?o }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	tt, ok := r.Bindings[0]["o"].(term.TripleTerm)
	if !ok {
		t.Fatal("expected TripleTerm")
	}
	lit := tt.Object().(term.Literal)
	if lit.Lexical() != "hello \"world\" \n\t" {
		t.Errorf("expected special chars preserved, got %q", lit.Lexical())
	}
}

func TestTripleTermVariableMatching(t *testing.T) {
	g := graph.NewGraph()
	g.Add(
		term.NewURIRefUnsafe("http://ex/s"),
		term.NewURIRefUnsafe("http://ex/p"),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral(42, term.WithDatatype(term.XSDInteger)),
		),
	)
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?s ?val WHERE {
			?s :p <<( :a :b ?val )>>
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	b := r.Bindings[0]
	if b["s"].(term.URIRef).Value() != "http://ex/s" {
		t.Error("subject not bound correctly")
	}
	if b["val"].(term.Literal).Lexical() != "42" {
		t.Error("inner variable not bound correctly")
	}
}

func TestNestedTripleTermVariable(t *testing.T) {
	g := graph.NewGraph()
	inner := term.NewTripleTerm(
		term.NewURIRefUnsafe("http://ex/x"),
		term.NewURIRefUnsafe("http://ex/y"),
		term.NewLiteral("z"),
	)
	outer := term.NewTripleTerm(
		term.NewURIRefUnsafe("http://ex/a"),
		term.NewURIRefUnsafe("http://ex/b"),
		inner,
	)
	g.Add(term.NewURIRefUnsafe("http://ex/s"), term.NewURIRefUnsafe("http://ex/has"), outer)
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?innerObj WHERE {
			:s :has <<( :a :b <<( :x :y ?innerObj )>> )>>
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	val := r.Bindings[0]["innerObj"]
	if val == nil {
		t.Fatal("innerObj is nil — nested triple term variable not bound")
	}
	if val.(term.Literal).Lexical() != "z" {
		t.Errorf("expected 'z', got %q", val.(term.Literal).Lexical())
	}
}

func TestAnnotationQuery(t *testing.T) {
	g := graph.NewGraph()
	s := term.NewURIRefUnsafe("http://ex/s")
	p := term.NewURIRefUnsafe("http://ex/p")
	o := term.NewURIRefUnsafe("http://ex/o")
	reifier := term.NewBNode("r1")
	rdfReifies := term.NewURIRefUnsafe("http://www.w3.org/1999/02/22-rdf-syntax-ns#reifies")
	g.Add(s, p, o)
	g.Add(reifier, rdfReifies, term.NewTripleTerm(s, p, o))
	g.Add(reifier, term.NewURIRefUnsafe("http://ex/source"), term.NewURIRefUnsafe("http://ex/web"))

	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?src WHERE {
			:s :p :o {| :source ?src |}
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	if r.Bindings[0]["src"].(term.URIRef).Value() != "http://ex/web" {
		t.Error("annotation source not matched")
	}
}

func TestTripleTermFunctions(t *testing.T) {
	g := graph.NewGraph()
	g.Add(
		term.NewURIRefUnsafe("http://ex/s"),
		term.NewURIRefUnsafe("http://ex/p"),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral(42, term.WithDatatype(term.XSDInteger)),
		),
	)
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?isT ?subj ?pred ?obj WHERE {
			?s :p ?tt .
			BIND(isTriple(?tt) AS ?isT)
			BIND(SUBJECT(?tt) AS ?subj)
			BIND(PREDICATE(?tt) AS ?pred)
			BIND(OBJECT(?tt) AS ?obj)
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	b := r.Bindings[0]
	if b["isT"].(term.Literal).Lexical() != "true" {
		t.Error("isTriple should be true")
	}
	if b["subj"].(term.URIRef).Value() != "http://ex/a" {
		t.Error("SUBJECT wrong")
	}
	if b["pred"].(term.URIRef).Value() != "http://ex/b" {
		t.Error("PREDICATE wrong")
	}
	if b["obj"].(term.Literal).Lexical() != "42" {
		t.Error("OBJECT wrong")
	}
}

func TestTripleTermInVALUES(t *testing.T) {
	g := graph.NewGraph()
	g.Add(
		term.NewURIRefUnsafe("http://ex/s"),
		term.NewURIRefUnsafe("http://ex/p"),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral(42, term.WithDatatype(term.XSDInteger)),
		),
	)
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?s WHERE {
			VALUES ?tt { <<( :a :b 42 )>> <<( :a :b 99 )>> }
			?s :p ?tt
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
}

func TestDirectionalLangTags(t *testing.T) {
	r, err := sparql.Query(graph.NewGraph(), `
		SELECT
			(LANGDIR("hello"@en--ltr) AS ?dir)
			(hasLANGDIR("hello"@en--ltr) AS ?has)
			(STRLANGDIR("abc", "ar", "rtl") AS ?built)
		WHERE {}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(r.Bindings))
	}
	b := r.Bindings[0]
	if b["dir"].(term.Literal).Lexical() != "ltr" {
		t.Error("LANGDIR wrong")
	}
	if b["has"].(term.Literal).Lexical() != "true" {
		t.Error("hasLANGDIR wrong")
	}
	lit := b["built"].(term.Literal)
	if lit.Language() != "ar" || lit.Dir() != "rtl" {
		t.Errorf("STRLANGDIR wrong: lang=%q dir=%q", lit.Language(), lit.Dir())
	}
}

func TestTripleTermCONSTRUCT(t *testing.T) {
	g := graph.NewGraph()
	g.Add(
		term.NewURIRefUnsafe("http://ex/s"),
		term.NewURIRefUnsafe("http://ex/p"),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral(42, term.WithDatatype(term.XSDInteger)),
		),
	)
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		CONSTRUCT {
			?s :annotated <<( ?s :p ?tt )>>
		} WHERE {
			?s :p ?tt
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if r.Graph == nil {
		t.Fatal("CONSTRUCT graph is nil")
	}
	count := 0
	for range r.Graph.Triples(nil, nil, nil) {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 triple, got %d", count)
	}
}

func TestUpdateWithReifiedTriples(t *testing.T) {
	ds := &sparql.Dataset{
		Default: graph.NewGraph(),
		NamedGraphs: map[string]*graph.Graph{
			"http://ex/g1": func() *graph.Graph {
				g := graph.NewGraph()
				g.Add(term.NewURIRefUnsafe("http://ex/a"), term.NewURIRefUnsafe("http://ex/b"), term.NewURIRefUnsafe("http://ex/c"))
				return g
			}(),
		},
	}
	err := sparql.Update(ds, `
		PREFIX : <http://ex/>
		INSERT { << ?s ?p ?o >> :from :g1 }
		WHERE { GRAPH :g1 { ?s ?p ?o } }
	`)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range ds.Default.Triples(nil, nil, nil) {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 triples (reifier + rdf:reifies), got %d", count)
	}
}

func TestVersionDirective(t *testing.T) {
	_, err := sparql.Parse(`VERSION "1.2" SELECT * WHERE { ?s ?p ?o }`)
	if err != nil {
		t.Errorf("VERSION directive should be accepted: %v", err)
	}
	_, err = sparql.Parse(`VERSION """1.2""" SELECT * WHERE { ?s ?p ?o }`)
	if err == nil {
		t.Error("triple-quoted VERSION should be rejected")
	}
}

func TestNegativeSyntax(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"literal subject in triple term expr", `SELECT * WHERE { BIND(<<( "lit" <http://p> <http://o> )>> AS ?t) }`},
		{"triple term subject in triple term expr", `SELECT * WHERE { BIND(<<( <<(<http://s> <http://p> <http://o>)>> <http://q> <http://z> )>> AS ?t) }`},
		{"reified triple in BIND", `SELECT * WHERE { ?s ?p ?o . BIND(<< ?s ?p ?o >> AS ?t) }`},
		{"nested aggregates", `SELECT (COUNT(COUNT(*)) AS ?c) WHERE {}`},
		{"duplicate VALUES vars", `SELECT * WHERE { VALUES (?a ?a) { (1 1) } }`},
		{"invalid lang direction", `SELECT ("foo"@en--foo AS ?v) WHERE {}`},
		{"bnode subject in ExprTripleTerm", `SELECT * WHERE { BIND(<<( _:b <http://p> <http://o> )>> AS ?t) }`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sparql.Parse(tc.query)
			if err == nil {
				t.Errorf("expected parse error for: %s", tc.query)
			}
		})
	}
}

func TestTripleTermDataSubjectMustBeIRI(t *testing.T) {
	ds := &sparql.Dataset{
		Default:     graph.NewGraph(),
		NamedGraphs: map[string]*graph.Graph{},
	}
	// Blank node as subject in triple term DATA should fail (rule [123])
	err := sparql.Update(ds, `INSERT DATA { <http://ex/s> <http://ex/p> <<( _:b <http://ex/q> <http://ex/o> )>> }`)
	if err == nil {
		t.Error("expected error: blank node subject in triple term data")
	}
	// IRI subject should succeed
	err = sparql.Update(ds, `INSERT DATA { <http://ex/s> <http://ex/p> <<( <http://ex/a> <http://ex/q> <http://ex/o> )>> }`)
	if err != nil {
		t.Fatalf("expected success for IRI subject in triple term data: %v", err)
	}
}

func TestEmptyGraphTripleTermQuery(t *testing.T) {
	r, err := sparql.Query(graph.NewGraph(), `SELECT ?s WHERE { <<( ?s ?p ?o )>> ?q ?z }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 0 {
		t.Errorf("expected 0 results from empty graph, got %d", len(r.Bindings))
	}
}

// --- Section 2: Property-Based Tests for Triple Term Functions ---

func TestTripleTermRoundTrip(t *testing.T) {
	g := graph.NewGraph()
	s := term.NewURIRefUnsafe("http://ex/s")
	p := term.NewURIRefUnsafe("http://ex/p")
	terms := []term.TripleTerm{
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewLiteral("hello"),
		),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/c"),
			term.NewURIRefUnsafe("http://ex/d"),
			term.NewLiteral(42, term.WithDatatype(term.XSDInteger)),
		),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/e"),
			term.NewURIRefUnsafe("http://ex/f"),
			term.NewTripleTerm(
				term.NewURIRefUnsafe("http://ex/g"),
				term.NewURIRefUnsafe("http://ex/h"),
				term.NewLiteral("nested"),
			),
		),
	}
	for _, tt := range terms {
		g.Add(s, p, tt)
	}
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?t ?rebuilt WHERE {
			:s :p ?t .
			FILTER(isTriple(?t))
			BIND(TRIPLE(SUBJECT(?t), PREDICATE(?t), OBJECT(?t)) AS ?rebuilt)
			FILTER(?t = ?rebuilt)
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != len(terms) {
		t.Errorf("expected %d round-trip matches, got %d", len(terms), len(r.Bindings))
	}
}

func TestIsTripleNegative(t *testing.T) {
	g := graph.NewGraph()
	s := term.NewURIRefUnsafe("http://ex/s")
	p := term.NewURIRefUnsafe("http://ex/p")
	g.Add(s, p, term.NewURIRefUnsafe("http://ex/a"))
	g.Add(s, p, term.NewLiteral("hello"))
	g.Add(s, p, term.NewLiteral(42, term.WithDatatype(term.XSDInteger)))
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?v WHERE {
			:s :p ?v .
			FILTER(isTriple(?v))
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 0 {
		t.Errorf("isTriple should be false for non-triple-terms, got %d results", len(r.Bindings))
	}
}

func TestAccessorsOnNonTripleTerm(t *testing.T) {
	g := graph.NewGraph()
	s := term.NewURIRefUnsafe("http://ex/s")
	p := term.NewURIRefUnsafe("http://ex/p")
	g.Add(s, p, term.NewURIRefUnsafe("http://ex/a"))
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT ?subj WHERE {
			:s :p ?v .
			BIND(SUBJECT(?v) AS ?subj)
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatal("expected 1 binding")
	}
	if r.Bindings[0]["subj"] != nil {
		t.Error("SUBJECT on URI should return nil")
	}
}

// --- Section 4: Large Graph Stress Tests ---

func TestLargeGraphTripleTerms(t *testing.T) {
	g := graph.NewGraph()
	for i := 0; i < 5000; i++ {
		g.Add(
			term.NewURIRefUnsafe(fmt.Sprintf("http://ex/s%d", i)),
			term.NewURIRefUnsafe("http://ex/p"),
			term.NewTripleTerm(
				term.NewURIRefUnsafe(fmt.Sprintf("http://ex/a%d", i%100)),
				term.NewURIRefUnsafe("http://ex/b"),
				term.NewLiteral(i),
			),
		)
	}
	r, err := sparql.Query(g, `
		PREFIX : <http://ex/>
		SELECT (COUNT(*) AS ?c) WHERE {
			?s :p <<( :a0 :b ?val )>>
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	count := r.Bindings[0]["c"].(term.Literal).Lexical()
	if count != "50" {
		t.Errorf("expected 50, got %s", count)
	}
}

func TestDeeplyNestedTripleTerms(t *testing.T) {
	var tmp_term term.Term = term.NewLiteral("leaf")
	for i := 0; i < 10; i++ {
		tmp_term = term.NewTripleTerm(
			term.NewURIRefUnsafe(fmt.Sprintf("http://ex/s%d", i)),
			term.NewURIRefUnsafe("http://ex/p"),
			tmp_term,
		)
	}
	g := graph.NewGraph()
	g.Add(term.NewURIRefUnsafe("http://ex/root"), term.NewURIRefUnsafe("http://ex/has"), tmp_term)

	r, err := sparql.Query(g, `SELECT ?o WHERE { ?s ?p ?o }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Bindings) != 1 {
		t.Fatal("expected 1 result")
	}
	n3 := r.Bindings[0]["o"].N3()
	if !strings.HasPrefix(n3, "<<( ") {
		t.Error("not a triple term")
	}
	if strings.Count(n3, "<<( ") != 10 {
		t.Errorf("expected 10 nesting levels, got %d", strings.Count(n3, "<<( "))
	}
}

// --- Section 5: N3/Serialization Round-Trip Tests ---

func TestTripleTermN3RoundTrip(t *testing.T) {
	cases := []term.TripleTerm{
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/s"),
			term.NewURIRefUnsafe("http://ex/p"),
			term.NewLiteral("hello \"world\""),
		),
		term.NewTripleTerm(
			term.NewURIRefUnsafe("http://ex/a"),
			term.NewURIRefUnsafe("http://ex/b"),
			term.NewTripleTerm(
				term.NewURIRefUnsafe("http://ex/nested"),
				term.NewURIRefUnsafe("http://ex/q"),
				term.NewLiteral("3.14", term.WithDatatype(term.XSDDecimal)),
			),
		),
	}
	for i, tt := range cases {
		n3 := tt.N3()
		g := graph.NewGraph()
		err := turtle.Parse(g, strings.NewReader(
			fmt.Sprintf("<http://ex/s> <http://ex/p> %s .", n3),
		))
		if err != nil {
			t.Fatalf("case %d: turtle parse failed: %v", i, err)
		}
		found := false
		for tr := range g.Triples(nil, nil, nil) {
			reparsed, ok := tr.Object.(term.TripleTerm)
			if !ok {
				t.Fatalf("case %d: not a triple term after reparse: %T", i, tr.Object)
			}
			if !tt.Equal(reparsed) {
				t.Errorf("case %d: round-trip failed:\n  original: %s\n  reparsed: %s", i, tt.N3(), reparsed.N3())
			}
			found = true
		}
		if !found {
			t.Fatalf("case %d: no triples found after reparse", i)
		}
	}
}

// --- Section 6: ResultsEqual with Triple Terms ---

func TestResultsEqualWithTripleTermBnodes(t *testing.T) {
	a := &sparql.Result{
		Type: "SELECT",
		Vars: []string{"x"},
		Bindings: []map[string]term.Term{{
			"x": term.NewTripleTerm(
				term.NewBNode("a1"),
				term.NewURIRefUnsafe("http://ex/p"),
				term.NewURIRefUnsafe("http://ex/o"),
			),
		}},
	}
	b := &sparql.Result{
		Type: "SELECT",
		Vars: []string{"x"},
		Bindings: []map[string]term.Term{{
			"x": term.NewTripleTerm(
				term.NewBNode("different_label"),
				term.NewURIRefUnsafe("http://ex/p"),
				term.NewURIRefUnsafe("http://ex/o"),
			),
		}},
	}
	if !sparql.ResultsEqual(a, b) {
		t.Error("results with different bnode labels in triple terms should be equal")
	}
}

// --- Section 7: SRJ/SRX Triple Term Parsing Edge Cases ---

func TestSRJMalformedTriple(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"missing predicate", `{"head":{"vars":["x"]},"results":{"bindings":[{"x":{"type":"triple","value":{"subject":{"type":"uri","value":"s"}}}}]}}`},
		{"null value", `{"head":{"vars":["x"]},"results":{"bindings":[{"x":{"type":"triple","value":null}}]}}`},
		{"empty object", `{"head":{"vars":["x"]},"results":{"bindings":[{"x":{"type":"triple","value":{}}}]}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := sparql.ParseSRJ(strings.NewReader(tc.input))
			if err != nil {
				return // parse error is acceptable
			}
			// If parsed, the binding should be nil (graceful degradation)
			if len(r.Bindings) > 0 && r.Bindings[0]["x"] != nil {
				t.Errorf("expected nil for malformed triple, got %v", r.Bindings[0]["x"])
			}
		})
	}
}
