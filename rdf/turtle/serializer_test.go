package turtle

import (
	"bytes"
	"cimgo/rdf/graph"
	"cimgo/rdf/namespace"
	"cimgo/rdf/term"
	"fmt"
	"strings"
	"testing"
)

// Ported from: test/test_serializers/test_serializer_turtle.py

func serializeTurtle(t *testing.T, g *graph.Graph) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Serialize(g, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestTurtleBasic(t *testing.T) {
	// Ported from: rdflib turtle serializer — basic subject-predicate-object
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	o := term.NewLiteral("hello")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Add(s, p, o)

	out := serializeTurtle(t, g)
	if !strings.Contains(out, "ex:s") {
		t.Errorf("expected prefixed subject, got:\n%s", out)
	}
	if !strings.Contains(out, `"hello"`) {
		t.Errorf("expected literal, got:\n%s", out)
	}
	if !strings.Contains(out, "@prefix ex:") {
		t.Errorf("expected prefix declaration, got:\n%s", out)
	}
}

func TestTurtleRDFTypeShorthand(t *testing.T) {
	// Ported from: rdflib turtle serializer — "a" shorthand for rdf:type
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/Alice")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Add(s, namespace.RDF.Type, term.NewURIRefUnsafe("http://example.org/Person"))

	out := serializeTurtle(t, g)
	if !strings.Contains(out, " a ") {
		t.Errorf("expected 'a' shorthand, got:\n%s", out)
	}
}

func TestTurtleMultiplePredicates(t *testing.T) {
	// Ported from: rdflib turtle serializer — semicolon separator
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p1, _ := term.NewURIRef("http://example.org/p1")
	p2, _ := term.NewURIRef("http://example.org/p2")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Add(s, p1, term.NewLiteral("a"))
	g.Add(s, p2, term.NewLiteral("b"))

	out := serializeTurtle(t, g)
	if !strings.Contains(out, ";") {
		t.Errorf("expected semicolon for multiple predicates, got:\n%s", out)
	}
}

func TestTurtleMultipleObjects(t *testing.T) {
	// Ported from: rdflib turtle serializer — comma separator
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Add(s, p, term.NewLiteral("a"))
	g.Add(s, p, term.NewLiteral("b"))

	out := serializeTurtle(t, g)
	if !strings.Contains(out, ",") {
		t.Errorf("expected comma for multiple objects, got:\n%s", out)
	}
}

func TestTurtleLiteralTypes(t *testing.T) {
	// Ported from: rdflib turtle serializer — literal shorthands
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))

	g.Add(s, p, term.NewLiteral(42))
	g.Add(s, p, term.NewLiteral(true))
	g.Add(s, p, term.NewLiteral("hello", term.WithLang("en")))

	out := serializeTurtle(t, g)
	if !strings.Contains(out, "42") {
		t.Errorf("expected integer shorthand, got:\n%s", out)
	}
	if !strings.Contains(out, "true") {
		t.Errorf("expected boolean shorthand, got:\n%s", out)
	}
	if !strings.Contains(out, `"hello"@en`) {
		t.Errorf("expected language tag, got:\n%s", out)
	}
}

func TestTurtlePrefixOnlyUsed(t *testing.T) {
	// Ported from: rdflib turtle serializer — only emit used prefixes
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Bind("unused", term.NewURIRefUnsafe("http://unused.org/"))
	g.Add(s, p, term.NewLiteral("v"))

	out := serializeTurtle(t, g)
	if strings.Contains(out, "unused") {
		t.Errorf("should not emit unused prefix, got:\n%s", out)
	}
}

func TestTurtleDeterministic(t *testing.T) {
	// Ported from: test/test_turtle_sort_issue613.py — deterministic output
	g := graph.NewGraph()
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	for i := 0; i < 5; i++ {
		s, _ := term.NewURIRef(fmt.Sprintf("http://example.org/s%d", i))
		p, _ := term.NewURIRef("http://example.org/p")
		g.Add(s, p, term.NewLiteral(fmt.Sprintf("v%d", i)))
	}

	out1 := serializeTurtle(t, g)
	out2 := serializeTurtle(t, g)
	if out1 != out2 {
		t.Errorf("output not deterministic:\n---\n%s\n---\n%s", out1, out2)
	}
}

func TestTurtleBase(t *testing.T) {
	// Ported from: rdflib turtle serializer — @base emission
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	g.Add(s, p, term.NewLiteral("v"))

	var buf bytes.Buffer
	err := Serialize(g, &buf, WithBase("http://example.org/"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "@base <http://example.org/>") {
		t.Errorf("expected @base, got:\n%s", buf.String())
	}
}

func TestTurtleInlineBNode(t *testing.T) {
	// Ported from: rdflib turtle serializer — blank node inlining as [ ... ]
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/p")
	name, _ := term.NewURIRef("http://example.org/name")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))

	bnode := term.NewBNode()
	g.Add(s, p, bnode)
	g.Add(bnode, name, term.NewLiteral("Alice"))

	out := serializeTurtle(t, g)
	if !strings.Contains(out, "[") || !strings.Contains(out, "]") {
		t.Errorf("expected inline blank node, got:\n%s", out)
	}
}

func TestTurtleCollection(t *testing.T) {
	// Ported from: rdflib turtle serializer — rdf:List as ( ... )
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/list")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))

	// Build list: (1 2 3)
	n1 := term.NewBNode()
	n2 := term.NewBNode()
	n3 := term.NewBNode()
	g.Add(s, p, n1)
	g.Add(n1, namespace.RDF.First, term.NewLiteral(1))
	g.Add(n1, namespace.RDF.Rest, n2)
	g.Add(n2, namespace.RDF.First, term.NewLiteral(2))
	g.Add(n2, namespace.RDF.Rest, n3)
	g.Add(n3, namespace.RDF.First, term.NewLiteral(3))
	g.Add(n3, namespace.RDF.Rest, namespace.RDF.Nil)

	out := serializeTurtle(t, g)
	if !strings.Contains(out, "( ") {
		t.Errorf("expected collection syntax, got:\n%s", out)
	}
}

func TestTurtleSortRDFTypeFirst(t *testing.T) {
	// Ported from: rdflib turtle serializer — rdf:type predicate comes first
	g := graph.NewGraph()
	s, _ := term.NewURIRef("http://example.org/s")
	p, _ := term.NewURIRef("http://example.org/name")
	g.Bind("ex", term.NewURIRefUnsafe("http://example.org/"))
	g.Add(s, p, term.NewLiteral("Alice"))
	g.Add(s, namespace.RDF.Type, term.NewURIRefUnsafe("http://example.org/Person"))

	out := serializeTurtle(t, g)
	aIdx := strings.Index(out, " a ")
	nameIdx := strings.Index(out, "ex:name")
	if aIdx < 0 || nameIdx < 0 || aIdx > nameIdx {
		t.Errorf("rdf:type should come before other predicates, got:\n%s", out)
	}
}
