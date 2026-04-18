package rdfxml

import (
	"cimgo/rdf/graph"
	"cimgo/rdf/namespace"
	"cimgo/rdf/term"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"
)

func termKey(t term.Term) string { return t.N3() }

// Serialize serializes a Graph to RDF/XML format.
// It groups triples by subject and emits typed node elements when rdf:type is present.
// If a subject has multiple rdf:type values, the first one with a valid QName becomes
// the element name; the rest are emitted as rdf:type property elements.
// Options: WithBase sets xml:base on the root element.
func Serialize(g *graph.Graph, w io.Writer, opts ...Option) error {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	base := cfg.base

	// Collect namespace prefixes
	nsMap := make(map[string]string) // namespace -> prefix
	nsMap[rdfNS] = "rdf"
	g.Namespaces()(func(prefix string, ns term.URIRef) bool {
		nsMap[ns.Value()] = prefix
		return true
	})

	// Group triples by subject
	subjects := make(map[string][]term.Triple)
	var subjectOrder []string
	g.Triples(nil, nil, nil)(func(t term.Triple) bool {
		sk := termKey(t.Subject)
		if _, exists := subjects[sk]; !exists {
			subjectOrder = append(subjectOrder, sk)
		}
		subjects[sk] = append(subjects[sk], t)
		return true
	})
	slices.Sort(subjectOrder)

	// Write XML header
	if _, err := fmt.Fprintln(w, `<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		return err
	}

	// rdf:RDF opening with namespace declarations
	if _, err := fmt.Fprintf(w, "<rdf:RDF"); err != nil {
		return err
	}
	var nsList []string
	for ns := range nsMap {
		nsList = append(nsList, ns)
	}
	slices.Sort(nsList)
	for _, ns := range nsList {
		prefix := nsMap[ns]
		if _, err := fmt.Fprintf(w, "\n   xmlns:%s=%s", prefix, xmlAttr(ns)); err != nil {
			return err
		}
	}
	if base != "" {
		if _, err := fmt.Fprintf(w, "\n   xml:base=%s", xmlAttr(base)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, ">"); err != nil {
		return err
	}

	// Write each subject
	for _, sk := range subjectOrder {
		triples := subjects[sk]
		if len(triples) == 0 {
			continue
		}
		subj := triples[0].Subject

		// Determine element type: first rdf:type with a valid QName becomes element name,
		// remaining rdf:type triples are kept as property elements.
		elemName := "rdf:Description"
		elemNameSet := false
		var remaining []term.Triple
		for _, t := range triples {
			if t.Predicate == namespace.RDF.Type {
				if u, ok := t.Object.(term.URIRef); ok {
					if !elemNameSet {
						qn := xmlQName(u.Value(), nsMap)
						if qn != "" {
							elemName = qn
							elemNameSet = true
							continue
						}
					}
				}
				// Additional rdf:type values become property elements
				remaining = append(remaining, t)
				continue
			}
			remaining = append(remaining, t)
		}

		// Sort remaining triples for determinism
		slices.SortFunc(remaining, func(a, b term.Triple) int {
			return strings.Compare(a.Predicate.N3()+a.Object.N3(), b.Predicate.N3()+b.Object.N3())
		})

		// Opening tag
		switch v := subj.(type) {
		case term.URIRef:
			if _, err := fmt.Fprintf(w, "  <%s rdf:about=%s>\n", elemName, xmlAttr(v.Value())); err != nil {
				return err
			}
		case term.BNode:
			if _, err := fmt.Fprintf(w, "  <%s rdf:nodeID=%s>\n", elemName, xmlAttr(v.Value())); err != nil {
				return err
			}
		}

		// Property elements
		for _, t := range remaining {
			predQN := xmlQName(t.Predicate.Value(), nsMap)
			if predQN == "" {
				predQN = t.Predicate.Value()
			}

			var err error
			switch obj := t.Object.(type) {
			case term.URIRef:
				_, err = fmt.Fprintf(w, "    <%s rdf:resource=%s/>\n", predQN, xmlAttr(obj.Value()))
			case term.BNode:
				_, err = fmt.Fprintf(w, "    <%s rdf:nodeID=%s/>\n", predQN, xmlAttr(obj.Value()))
			case term.Literal:
				if obj.Language() != "" {
					_, err = fmt.Fprintf(w, "    <%s xml:lang=%s>%s</%s>\n", predQN, xmlAttr(obj.Language()), xmlEscape(obj.Lexical()), predQN)
				} else if obj.Datatype() != (term.URIRef{}) && obj.Datatype() != term.XSDString {
					_, err = fmt.Fprintf(w, "    <%s rdf:datatype=%s>%s</%s>\n", predQN, xmlAttr(obj.Datatype().Value()), xmlEscape(obj.Lexical()), predQN)
				} else {
					_, err = fmt.Fprintf(w, "    <%s>%s</%s>\n", predQN, xmlEscape(obj.Lexical()), predQN)
				}
			}
			if err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(w, "  </%s>\n", elemName); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w, "</rdf:RDF>")
	return err
}

func xmlQName(uri string, nsMap map[string]string) string {
	bestNS := ""
	bestPrefix := ""
	for ns, prefix := range nsMap {
		if strings.HasPrefix(uri, ns) && len(ns) > len(bestNS) {
			local := uri[len(ns):]
			if local != "" && !strings.ContainsAny(local, "/#") {
				bestNS = ns
				bestPrefix = prefix
			}
		}
	}
	if bestNS != "" {
		return bestPrefix + ":" + uri[len(bestNS):]
	}
	return ""
}

// xmlAttr returns an XML-escaped, double-quoted attribute value.
func xmlAttr(s string) string {
	var b strings.Builder
	// xml.EscapeText writing to strings.Builder cannot fail, so we ignore the error.
	_ = xml.EscapeText(&b, []byte(s))
	return `"` + b.String() + `"`
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
