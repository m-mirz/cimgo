package sparql

import (
	"cimgo/rdf/term"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// queryStartTimeKey is used to pass the query start time through the prefixes map.
const queryStartTimeKey = "__query_start_time__"

// regexCache caches compiled regular expressions to avoid recompilation per row.
var regexCache sync.Map // pattern string → *regexp.Regexp

func cachedRegexpCompile(pattern string) (*regexp.Regexp, error) {
	if v, ok := regexCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexCache.Store(pattern, re)
	return re, nil
}

// evalFunc evaluates a SPARQL built-in function.
// Ported from: rdflib.plugins.sparql.operators
func evalFunc(name string, args []Expr, bindings map[string]term.Term, prefixes map[string]string) term.Term {
	evalArgs := func() []term.Term {
		var vals []term.Term
		for _, a := range args {
			vals = append(vals, evalExpr(a, bindings, prefixes))
		}
		return vals
	}

	switch name {
	// Term constructors
	case "BOUND":
		if len(args) == 1 {
			if v, ok := args[0].(*VarExpr); ok {
				_, exists := bindings[v.Name]
				return term.NewLiteral(exists)
			}
		}
		return term.NewLiteral(false)

	case "ISIRI", "ISURI":
		vals := evalArgs()
		if len(vals) == 1 {
			_, ok := vals[0].(term.URIRef)
			return term.NewLiteral(ok)
		}
	case "ISBLANK":
		vals := evalArgs()
		if len(vals) == 1 {
			_, ok := vals[0].(term.BNode)
			return term.NewLiteral(ok)
		}
	case "ISLITERAL":
		vals := evalArgs()
		if len(vals) == 1 {
			_, ok := vals[0].(term.Literal)
			return term.NewLiteral(ok)
		}
	case "ISNUMERIC":
		vals := evalArgs()
		if len(vals) == 1 {
			if l, ok := vals[0].(term.Literal); ok {
				dt := l.Datatype()
				return term.NewLiteral(dt == term.XSDInteger || dt == term.XSDFloat || dt == term.XSDDouble || dt == term.XSDDecimal)
			}
		}
		return term.NewLiteral(false)

	// String functions
	case "STR":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			switch v := vals[0].(type) {
			case term.URIRef:
				return term.NewLiteral(v.Value())
			case term.Literal:
				return term.NewLiteral(v.Lexical())
			default:
				return term.NewLiteral(vals[0].String())
			}
		}
	case "STRLEN":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(utf8.RuneCountInString(termString(vals[0])))
		}
	case "SUBSTR":
		vals := evalArgs()
		if len(vals) < 1 {
			return nil
		}
		s := termString(vals[0])
		runes := []rune(s)
		if len(vals) >= 2 {
			start := int(toFloat64(vals[1])) - 1 // SPARQL is 1-based
			if start < 0 {
				start = 0
			}
			if start >= len(runes) {
				return stringResult("", vals[0])
			}
			if len(vals) >= 3 {
				length := int(toFloat64(vals[2]))
				end := start + length
				if end > len(runes) {
					end = len(runes)
				}
				return stringResult(string(runes[start:end]), vals[0])
			}
			return stringResult(string(runes[start:]), vals[0])
		}
	case "UCASE":
		vals := evalArgs()
		if len(vals) == 1 {
			return stringResult(strings.ToUpper(termString(vals[0])), vals[0])
		}
	case "LCASE":
		vals := evalArgs()
		if len(vals) == 1 {
			return stringResult(strings.ToLower(termString(vals[0])), vals[0])
		}
	case "STRSTARTS":
		vals := evalArgs()
		if len(vals) == 2 {
			return term.NewLiteral(strings.HasPrefix(termString(vals[0]), termString(vals[1])))
		}
	case "STRENDS":
		vals := evalArgs()
		if len(vals) == 2 {
			return term.NewLiteral(strings.HasSuffix(termString(vals[0]), termString(vals[1])))
		}
	case "CONTAINS":
		vals := evalArgs()
		if len(vals) == 2 {
			return term.NewLiteral(strings.Contains(termString(vals[0]), termString(vals[1])))
		}
	case "CONCAT":
		vals := evalArgs()
		var sb strings.Builder
		// Track language and direction: preserve only if ALL args match
		var commonLang *string
		var commonDir *string
		hasError := false
		for _, v := range vals {
			if v == nil {
				hasError = true
				continue
			}
			// CONCAT requires string-compatible arguments
			if !isStringLiteral(v) {
				hasError = true
				continue
			}
			sb.WriteString(termString(v))
			if l, ok := v.(term.Literal); ok {
				lang := l.Language()
				dir := l.Dir()
				if commonLang == nil {
					commonLang = &lang
				} else if *commonLang != lang {
					empty := ""
					commonLang = &empty
				}
				if commonDir == nil {
					commonDir = &dir
				} else if *commonDir != dir {
					empty := ""
					commonDir = &empty
				}
			}
		}
		if hasError {
			return nil
		}
		var opts []term.LiteralOption
		if commonLang != nil && *commonLang != "" && commonDir != nil && *commonDir != "" {
			// All have same lang AND same dir
			opts = append(opts, term.WithLang(*commonLang))
			opts = append(opts, term.WithDir(*commonDir))
		} else if commonLang != nil && *commonLang != "" && (commonDir == nil || *commonDir == "") {
			// Check if any had dir — if some had dir and some didn't, drop lang
			noneHaveDir := true
			for _, v := range vals {
				if l, ok := v.(term.Literal); ok {
					if l.Dir() != "" {
						noneHaveDir = false
						break
					}
				}
			}
			if noneHaveDir {
				opts = append(opts, term.WithLang(*commonLang))
			}
		}
		return term.NewLiteral(sb.String(), opts...)
	case "REGEX":
		vals := evalArgs()
		if len(vals) >= 2 {
			pattern := termString(vals[1])
			flags := ""
			if len(vals) >= 3 {
				flags = termString(vals[2])
			}
			if strings.Contains(flags, "i") {
				pattern = "(?i)" + pattern
			}
			re, err := cachedRegexpCompile(pattern)
			if err != nil {
				return term.NewLiteral(false)
			}
			return term.NewLiteral(re.MatchString(termString(vals[0])))
		}
	case "REPLACE":
		vals := evalArgs()
		if len(vals) >= 3 {
			// REPLACE requires string literal input
			if l, ok := vals[0].(term.Literal); ok {
				if isNumericDatatype(l.Datatype()) {
					return nil // type error
				}
			} else {
				return nil // non-literal
			}
			pattern := termString(vals[1])
			replacement := termString(vals[2])
			flags := ""
			if len(vals) >= 4 {
				flags = termString(vals[3])
			}
			if strings.Contains(flags, "i") {
				pattern = "(?i)" + pattern
			}
			re, err := cachedRegexpCompile(pattern)
			if err != nil {
				return vals[0]
			}
			return stringResult(re.ReplaceAllString(termString(vals[0]), replacement), vals[0])
		}

	// Term accessors
	case "LANG":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			if l, ok := vals[0].(term.Literal); ok {
				return term.NewLiteral(l.Language())
			}
			return nil // type error for non-literals
		}
		return term.NewLiteral("")
	case "DATATYPE":
		vals := evalArgs()
		if len(vals) == 1 {
			if l, ok := vals[0].(term.Literal); ok {
				return l.Datatype()
			}
		}

	// Numeric
	case "ABS":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(math.Abs(toFloat64(vals[0])))
		}
	case "ROUND":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(math.Round(toFloat64(vals[0])))
		}
	case "CEIL":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(math.Ceil(toFloat64(vals[0])))
		}
	case "FLOOR":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(math.Floor(toFloat64(vals[0])))
		}

	// Hash
	case "MD5":
		vals := evalArgs()
		if len(vals) == 1 {
			h := md5.Sum([]byte(termString(vals[0])))
			return term.NewLiteral(fmt.Sprintf("%x", h))
		}
	case "SHA1":
		vals := evalArgs()
		if len(vals) == 1 {
			h := sha1.Sum([]byte(termString(vals[0])))
			return term.NewLiteral(fmt.Sprintf("%x", h))
		}
	case "SHA256":
		vals := evalArgs()
		if len(vals) == 1 {
			h := sha256.Sum256([]byte(termString(vals[0])))
			return term.NewLiteral(fmt.Sprintf("%x", h))
		}

	// Conditional
	case "IF":
		if len(args) == 3 {
			cond := evalExpr(args[0], bindings, prefixes)
			if cond == nil {
				return nil // error in condition propagates
			}
			if effectiveBooleanValue(cond) {
				return evalExpr(args[1], bindings, prefixes)
			}
			return evalExpr(args[2], bindings, prefixes)
		}
	case "COALESCE":
		for _, a := range args {
			v := evalExpr(a, bindings, prefixes)
			if v != nil {
				return v
			}
		}
		return nil
	case "LANGMATCHES":
		vals := evalArgs()
		if len(vals) == 2 {
			tag := strings.ToLower(termString(vals[0]))
			range_ := strings.ToLower(termString(vals[1]))
			if range_ == "*" {
				return term.NewLiteral(tag != "")
			}
			return term.NewLiteral(tag == range_ || strings.HasPrefix(tag, range_+"-"))
		}
		return term.NewLiteral(false)
	case "SAMETERM":
		vals := evalArgs()
		if len(vals) == 2 && vals[0] != nil && vals[1] != nil {
			return term.NewLiteral(vals[0].N3() == vals[1].N3())
		}
		return term.NewLiteral(false)

	// String constructors
	case "STRLANG":
		vals := evalArgs()
		if len(vals) == 2 && vals[0] != nil && vals[1] != nil {
			// STRLANG requires a simple literal (no language, no datatype other than xsd:string)
			l, ok := vals[0].(term.Literal)
			if !ok {
				return nil // type error: non-literal
			}
			if l.Language() != "" {
				return nil // type error
			}
			dt := l.Datatype()
			if dt != term.XSDString && dt.Value() != "" {
				return nil // type error: has non-string datatype
			}
			lang := termString(vals[1])
			if lang == "" {
				return nil // empty language tag is an error
			}
			return term.NewLiteral(l.Lexical(), term.WithLang(lang))
		}
	case "STRDT":
		vals := evalArgs()
		if len(vals) == 2 && vals[0] != nil {
			// STRDT requires a simple literal (no language, xsd:string or no datatype)
			if !isStringLiteral(vals[0]) {
				return nil // type error: non-string input
			}
			if l, ok := vals[0].(term.Literal); ok {
				if l.Language() != "" {
					return nil // type error: has language tag
				}
			}
			if u, ok := vals[1].(term.URIRef); ok {
				return term.NewLiteral(termString(vals[0]), term.WithDatatype(u))
			}
		}
	case "STRBEFORE":
		vals := evalArgs()
		if len(vals) == 2 && vals[0] != nil && vals[1] != nil {
			if !isStringLiteral(vals[0]) || !isStringLiteral(vals[1]) {
				return nil
			}
			if !strArgCompatible(vals[0], vals[1]) {
				return nil
			}
			s := termString(vals[0])
			arg := termString(vals[1])
			if arg == "" {
				return stringResult("", vals[0])
			}
			idx := strings.Index(s, arg)
			if idx < 0 {
				return term.NewLiteral("")
			}
			return stringResult(s[:idx], vals[0])
		}
	case "STRAFTER":
		vals := evalArgs()
		if len(vals) == 2 && vals[0] != nil && vals[1] != nil {
			if !isStringLiteral(vals[0]) || !isStringLiteral(vals[1]) {
				return nil
			}
			if !strArgCompatible(vals[0], vals[1]) {
				return nil
			}
			s := termString(vals[0])
			arg := termString(vals[1])
			if arg == "" {
				return stringResult(s, vals[0])
			}
			idx := strings.Index(s, arg)
			if idx < 0 {
				return term.NewLiteral("")
			}
			return stringResult(s[idx+len(arg):], vals[0])
		}
	case "ENCODE_FOR_URI":
		vals := evalArgs()
		if len(vals) == 1 {
			return term.NewLiteral(encodeForURI(termString(vals[0])))
		}
	case "IRI", "URI":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			var s string
			if u, ok := vals[0].(term.URIRef); ok {
				s = u.Value()
			} else {
				s = termString(vals[0])
			}
			// Resolve relative URI against base
			if base, ok := prefixes[baseURIKey]; ok && !strings.Contains(s, ":") {
				s = base + s
			}
			return term.NewURIRefUnsafe(s)
		}
	case "BNODE":
		if len(args) == 0 {
			return term.NewBNode("") // unique each call
		}
		vals := evalArgs()
		if len(vals) == 1 {
			// BNODE(str): same str → same bnode within a query
			key := termString(vals[0])
			// Use a deterministic bnode label based on the input
			return term.NewBNode("bnode_" + key)
		}

	// Date/time functions
	case "NOW":
		// Per SPARQL 1.1 spec §17.4.5.1, NOW() must return the same value
		// throughout a single query evaluation.
		nowStr := prefixes[queryStartTimeKey]
		if nowStr == "" {
			nowStr = timeNow()
		}
		return term.NewLiteral(nowStr, term.WithDatatype(term.NewURIRefUnsafe("http://www.w3.org/2001/XMLSchema#dateTime")))
	case "YEAR":
		vals := evalArgs()
		if len(vals) == 1 {
			if y, ok := extractDatePart(termString(vals[0]), "year"); ok {
				return term.NewLiteral(y, term.WithDatatype(term.XSDInteger))
			}
		}
	case "MONTH":
		vals := evalArgs()
		if len(vals) == 1 {
			if m, ok := extractDatePart(termString(vals[0]), "month"); ok {
				return term.NewLiteral(m, term.WithDatatype(term.XSDInteger))
			}
		}
	case "DAY":
		vals := evalArgs()
		if len(vals) == 1 {
			if d, ok := extractDatePart(termString(vals[0]), "day"); ok {
				return term.NewLiteral(d, term.WithDatatype(term.XSDInteger))
			}
		}
	case "HOURS":
		vals := evalArgs()
		if len(vals) == 1 {
			if h, ok := extractDatePart(termString(vals[0]), "hours"); ok {
				return term.NewLiteral(h, term.WithDatatype(term.XSDInteger))
			}
		}
	case "MINUTES":
		vals := evalArgs()
		if len(vals) == 1 {
			if m, ok := extractDatePart(termString(vals[0]), "minutes"); ok {
				return term.NewLiteral(m, term.WithDatatype(term.XSDInteger))
			}
		}
	case "SECONDS":
		vals := evalArgs()
		if len(vals) == 1 {
			if s, ok := extractDatePart(termString(vals[0]), "seconds"); ok {
				return term.NewLiteral(s, term.WithDatatype(term.XSDDecimal))
			}
		}
	case "TIMEZONE":
		vals := evalArgs()
		if len(vals) == 1 {
			if tz, ok := extractTimezone(termString(vals[0])); ok {
				return term.NewLiteral(tz, term.WithDatatype(term.NewURIRefUnsafe("http://www.w3.org/2001/XMLSchema#dayTimeDuration")))
			}
		}
	case "TZ":
		vals := evalArgs()
		if len(vals) == 1 {
			if tz, ok := extractTZ(termString(vals[0])); ok {
				return term.NewLiteral(tz)
			}
		}

	// Hash
	case "SHA384":
		vals := evalArgs()
		if len(vals) == 1 {
			h := sha512.Sum384([]byte(termString(vals[0])))
			return term.NewLiteral(fmt.Sprintf("%x", h))
		}
	case "SHA512":
		vals := evalArgs()
		if len(vals) == 1 {
			h := sha512.Sum512([]byte(termString(vals[0])))
			return term.NewLiteral(fmt.Sprintf("%x", h))
		}

	// Random/UUID
	case "RAND":
		return term.NewLiteral(randFloat(), term.WithDatatype(term.XSDDouble))
	case "UUID":
		return term.NewURIRefUnsafe("urn:uuid:" + newUUID())
	case "STRUUID":
		return term.NewLiteral(newUUID())

	// Triple term functions (SPARQL 1.2)
	case "ISTRIPLE":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			_, ok := vals[0].(term.TripleTerm)
			return term.NewLiteral(ok)
		}
		return term.NewLiteral(false)
	case "TRIPLE":
		vals := evalArgs()
		if len(vals) == 3 && vals[0] != nil && vals[1] != nil && vals[2] != nil {
			subj, ok := vals[0].(term.Subject)
			if !ok {
				return nil
			}
			pred, ok := vals[1].(term.URIRef)
			if !ok {
				return nil
			}
			return term.NewTripleTerm(subj, pred, vals[2])
		}
	case "SUBJECT":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			if tt, ok := vals[0].(term.TripleTerm); ok {
				return tt.Subject()
			}
		}
	case "PREDICATE":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			if tt, ok := vals[0].(term.TripleTerm); ok {
				return tt.Predicate()
			}
		}
	case "OBJECT":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			if tt, ok := vals[0].(term.TripleTerm); ok {
				return tt.Object()
			}
		}

	// Language direction functions (SPARQL 1.2)
	case "LANGDIR":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			if l, ok := vals[0].(term.Literal); ok {
				return term.NewLiteral(l.Dir())
			}
			return nil // type error for non-literals
		}
		return term.NewLiteral("")
	case "HASLANG":
		vals := evalArgs()
		if len(vals) == 1 {
			if l, ok := vals[0].(term.Literal); ok {
				return term.NewLiteral(l.Language() != "")
			}
		}
		return term.NewLiteral(false)
	case "HASLANGDIR":
		vals := evalArgs()
		if len(vals) == 1 {
			if l, ok := vals[0].(term.Literal); ok {
				return term.NewLiteral(l.Dir() != "")
			}
		}
		return term.NewLiteral(false)
	case "STRLANGDIR":
		vals := evalArgs()
		if len(vals) == 3 && vals[0] != nil && vals[1] != nil && vals[2] != nil {
			// STRLANGDIR requires a simple literal (like STRLANG)
			l, ok := vals[0].(term.Literal)
			if !ok {
				return nil // non-literal
			}
			if l.Language() != "" {
				return nil // type error
			}
			dt := l.Datatype()
			if dt != term.XSDString && dt.Value() != "" {
				return nil // type error
			}
			lang := termString(vals[1])
			dir := termString(vals[2])
			// Direction must be exactly "ltr" or "rtl" (case-sensitive)
			if dir != "ltr" && dir != "rtl" {
				return nil // invalid or empty direction
			}
			// Per RDF 1.2, dirLangString requires a language tag
			if lang == "" {
				return nil // empty lang with dir is invalid
			}
			return term.NewLiteral(l.Lexical(), term.WithLang(lang), term.WithDir(dir))
		}

	// Cast functions
	case "XSD:BOOLEAN", "XSD:INTEGER", "XSD:FLOAT", "XSD:DOUBLE", "XSD:DECIMAL", "XSD:STRING":
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			return castXSD(name, vals[0])
		}
	}

	// Try cast with full IRI
	if strings.HasPrefix(name, "HTTP://WWW.W3.ORG/2001/XMLSCHEMA#") {
		vals := evalArgs()
		if len(vals) == 1 && vals[0] != nil {
			localName := strings.ToUpper(name[len("HTTP://WWW.W3.ORG/2001/XMLSCHEMA#"):])
			return castXSD("XSD:"+localName, vals[0])
		}
	}

	return nil
}

// --- Helpers ---

func effectiveBooleanValue(t term.Term) bool {
	if t == nil {
		return false
	}
	if l, ok := t.(term.Literal); ok {
		switch l.Datatype() {
		case term.XSDBoolean:
			return l.Lexical() == "true" || l.Lexical() == "1"
		case term.XSDInteger, term.XSDInt, term.XSDLong:
			v, _ := strconv.ParseInt(l.Lexical(), 10, 64)
			return v != 0
		case term.XSDFloat, term.XSDDouble, term.XSDDecimal:
			v, _ := strconv.ParseFloat(l.Lexical(), 64)
			return v != 0
		case term.XSDString:
			return l.Lexical() != ""
		default:
			return l.Lexical() != ""
		}
	}
	return true
}

func toFloat64(t term.Term) float64 {
	if t == nil {
		return 0
	}
	if l, ok := t.(term.Literal); ok {
		f, _ := strconv.ParseFloat(l.Lexical(), 64)
		return f
	}
	return 0
}

func isIntegral(t term.Term) bool {
	if l, ok := t.(term.Literal); ok {
		return l.Datatype() == term.XSDInteger || l.Datatype() == term.XSDInt || l.Datatype() == term.XSDLong
	}
	return false
}

func termString(t term.Term) string {
	if t == nil {
		return ""
	}
	return t.String()
}

// stringResult creates a literal preserving language/datatype from the source term.
func stringResult(s string, source term.Term) term.Literal {
	if l, ok := source.(term.Literal); ok {
		if lang := l.Language(); lang != "" {
			return term.NewLiteral(s, term.WithLang(lang))
		}
		if dt := l.Datatype(); dt != term.XSDString {
			return term.NewLiteral(s, term.WithDatatype(dt))
		}
	}
	return term.NewLiteral(s)
}

// isStringLiteral checks if a term is a string-type literal (plain, xsd:string, or lang-tagged).
func isStringLiteral(t term.Term) bool {
	l, ok := t.(term.Literal)
	if !ok {
		return false
	}
	dt := l.Datatype()
	return dt == term.XSDString || l.Language() != "" || dt.Value() == ""
}

// strArgCompatible checks if two string arguments are type-compatible for
// STRBEFORE/STRAFTER per SPARQL spec. Compatible if:
// - Both are simple/xsd:string literals (no lang)
// - Both have the same language tag
// - Second arg is simple/xsd:string (no lang)
func strArgCompatible(a, b term.Term) bool {
	la, aLit := a.(term.Literal)
	lb, bLit := b.(term.Literal)
	if !aLit || !bLit {
		return false
	}
	aLang := la.Language()
	bLang := lb.Language()
	// If second arg has a language, first must have the same language
	if bLang != "" {
		return strings.EqualFold(aLang, bLang)
	}
	// Second arg is simple — compatible with anything
	return true
}

// encodeForURI implements SPARQL ENCODE_FOR_URI per §17.4.3.14.
// Only RFC 3986 unreserved characters (A-Z a-z 0-9 - _ . ~) are left unencoded.
func encodeForURI(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 3) // worst case
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func timeNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func extractDatePart(dt, part string) (string, bool) {
	// Parse ISO 8601 datetime: 2011-01-10T14:45:13.815-05:00
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", dt)
		if err != nil {
			t, err = time.Parse("2006-01-02", dt)
			if err != nil {
				return "", false
			}
		}
	}
	switch part {
	case "year":
		return strconv.Itoa(t.Year()), true
	case "month":
		return strconv.Itoa(int(t.Month())), true
	case "day":
		return strconv.Itoa(t.Day()), true
	case "hours":
		return strconv.Itoa(t.Hour()), true
	case "minutes":
		return strconv.Itoa(t.Minute()), true
	case "seconds":
		sec := float64(t.Second()) + float64(t.Nanosecond())/1e9
		if t.Nanosecond() == 0 {
			return fmt.Sprintf("%d", t.Second()), true
		}
		return fmt.Sprintf("%g", sec), true
	}
	return "", false
}

func extractTimezone(dt string) (string, bool) {
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		return "", false
	}
	_, offset := t.Zone()
	if offset == 0 {
		return "PT0S", true
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	sign := ""
	if hours < 0 {
		sign = "-"
		hours = -hours
		minutes = -minutes
	}
	if minutes == 0 {
		return fmt.Sprintf("%sPT%dH", sign, hours), true
	}
	return fmt.Sprintf("%sPT%dH%dM", sign, hours, minutes), true
}

func extractTZ(dt string) (string, bool) {
	// Return timezone string like "Z", "-05:00", etc.
	if strings.HasSuffix(dt, "Z") {
		return "Z", true
	}
	// Look for +HH:MM or -HH:MM at end
	if len(dt) >= 6 {
		tz := dt[len(dt)-6:]
		if (tz[0] == '+' || tz[0] == '-') && tz[3] == ':' {
			return tz, true
		}
	}
	return "", true // no timezone info
}

func randFloat() float64 {
	return rand.Float64() // math/rand/v2: goroutine-safe global source
}

func newUUID() string {
	return uuid.New().String()
}

func castXSD(name string, val term.Term) term.Term {
	lit, isLit := val.(term.Literal)
	_, isURI := val.(term.URIRef)

	switch name {
	case "XSD:BOOLEAN":
		if isURI {
			return nil // can't cast URI to boolean
		}
		if !isLit {
			return nil
		}
		s := lit.Lexical()
		dt := lit.Datatype()
		if dt == term.XSDBoolean {
			// Normalize: "0"/"1" → "false"/"true"
			return term.NewLiteral(effectiveBooleanValue(val))
		}
		if isNumericDatatype(dt) {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil
			}
			return term.NewLiteral(f != 0)
		}
		// String/plain literal
		switch strings.ToLower(s) {
		case "true", "1":
			return term.NewLiteral(true)
		case "false", "0":
			return term.NewLiteral(false)
		default:
			return nil // can't cast arbitrary string to boolean
		}

	case "XSD:INTEGER":
		if !isLit {
			return nil
		}
		s := lit.Lexical()
		dt := lit.Datatype()
		if dt == term.XSDBoolean {
			if s == "true" || s == "1" {
				return term.NewLiteral(1, term.WithDatatype(term.XSDInteger))
			}
			return term.NewLiteral(0, term.WithDatatype(term.XSDInteger))
		}
		if isNumericDatatype(dt) {
			// From numeric: truncate to integer
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if math.IsNaN(f) || math.IsInf(f, 0) || f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
					return nil
				}
				return term.NewLiteral(int64(f), term.WithDatatype(term.XSDInteger))
			}
		}
		// From string/plain: must be a valid integer lexical form
		if _, err := strconv.ParseInt(strings.TrimLeft(s, "+"), 10, 64); err == nil {
			return term.NewLiteral(s, term.WithDatatype(term.XSDInteger))
		}
		return nil

	case "XSD:FLOAT":
		if !isLit {
			return nil
		}
		s := lit.Lexical()
		if lit.Datatype() == term.XSDBoolean {
			if s == "true" || s == "1" {
				s = "1.0"
			} else {
				s = "0.0"
			}
		}
		if f, err := strconv.ParseFloat(s, 32); err == nil {
			return term.NewLiteral(strconv.FormatFloat(float64(float32(f)), 'E', -1, 32), term.WithDatatype(term.XSDFloat))
		}
		return nil

	case "XSD:DOUBLE":
		if !isLit {
			return nil
		}
		s := lit.Lexical()
		if lit.Datatype() == term.XSDBoolean {
			if s == "true" || s == "1" {
				s = "1.0"
			} else {
				s = "0.0"
			}
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return term.NewLiteral(strconv.FormatFloat(f, 'E', -1, 64), term.WithDatatype(term.XSDDouble))
		}
		return nil

	case "XSD:DECIMAL":
		if !isLit {
			return nil
		}
		s := lit.Lexical()
		dt := lit.Datatype()
		if dt == term.XSDBoolean {
			if effectiveBooleanValue(val) {
				return term.NewLiteral("1.0", term.WithDatatype(term.XSDDecimal))
			}
			return term.NewLiteral("0.0", term.WithDatatype(term.XSDDecimal))
		}
		// Reject scientific notation strings (not valid xsd:decimal)
		if !isNumericDatatype(dt) && strings.ContainsAny(s, "eE") {
			return nil
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return term.NewLiteral(formatDecimal(f), term.WithDatatype(term.XSDDecimal))
		}
		return nil

	case "XSD:STRING":
		if isURI {
			u := val.(term.URIRef)
			return term.NewLiteral(u.Value(), term.WithDatatype(term.XSDString))
		}
		if !isLit {
			return nil
		}
		// Canonical string representation per datatype
		dt := lit.Datatype()
		s := lit.Lexical()
		if dt == term.XSDBoolean {
			if effectiveBooleanValue(val) {
				s = "true"
			} else {
				s = "false"
			}
		} else if dt == term.XSDInteger || dt == term.XSDInt || dt == term.XSDLong {
			if v, err := strconv.ParseInt(s, 10, 64); err == nil {
				s = strconv.FormatInt(v, 10)
			}
		} else if dt == term.XSDDecimal {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if f == float64(int64(f)) {
					s = strconv.FormatInt(int64(f), 10)
				} else {
					s = strconv.FormatFloat(f, 'f', -1, 64)
				}
			}
		} else if dt == term.XSDDouble || dt == term.XSDFloat {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if f == float64(int64(f)) && f != 0 {
					s = strconv.FormatInt(int64(f), 10)
				} else if f == 0 {
					s = "0"
				} else {
					s = strconv.FormatFloat(f, 'f', -1, 64)
				}
			}
		}
		return term.NewLiteral(s, term.WithDatatype(term.XSDString))
	}
	return nil
}

// termTypeOrder returns a numeric order for term types per SPARQL ordering:
// Blanks < IRIs < Literals < TripleTerms
func termTypeOrder(t term.Term) int {
	switch t.(type) {
	case term.BNode:
		return 0
	case term.URIRef:
		return 1
	case term.Literal:
		return 2
	case term.TripleTerm:
		return 3
	}
	return 4
}

func compareTermValues(a, b term.Term) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Different term types: compare by type order
	aOrder := termTypeOrder(a)
	bOrder := termTypeOrder(b)
	if aOrder != bOrder {
		return aOrder - bOrder
	}

	// Same term type
	la, okA := a.(term.Literal)
	lb, okB := b.(term.Literal)
	if okA && okB {
		fa, errA := strconv.ParseFloat(la.Lexical(), 64)
		fb, errB := strconv.ParseFloat(lb.Lexical(), 64)
		if errA == nil && errB == nil && isNumericDatatype(la.Datatype()) && isNumericDatatype(lb.Datatype()) {
			if math.IsNaN(fa) || math.IsNaN(fb) {
				return strings.Compare(a.N3(), b.N3())
			}
			if fa < fb {
				return -1
			}
			if fa > fb {
				return 1
			}
			return 0
		}
		// Date/dateTime comparison
		if isDateDatatype(la.Datatype()) && isDateDatatype(lb.Datatype()) {
			if ta, tb, ok := parseDatePair(la, lb); ok {
				if ta.Before(tb) {
					return -1
				}
				if ta.After(tb) {
					return 1
				}
				return 0
			}
		}
	}

	// URIs: compare by value, not N3 (to avoid angle bracket interference)
	uA, aIsURI := a.(term.URIRef)
	uB, bIsURI := b.(term.URIRef)
	if aIsURI && bIsURI {
		return strings.Compare(uA.Value(), uB.Value())
	}

	// Triple terms: compare component by component
	ttA, aIsTT := a.(term.TripleTerm)
	ttB, bIsTT := b.(term.TripleTerm)
	if aIsTT && bIsTT {
		if c := compareTermValues(ttA.Subject(), ttB.Subject()); c != 0 {
			return c
		}
		if c := compareTermValues(ttA.Predicate(), ttB.Predicate()); c != 0 {
			return c
		}
		return compareTermValues(ttA.Object(), ttB.Object())
	}

	return strings.Compare(a.N3(), b.N3())
}

func isDateDatatype(dt term.URIRef) bool {
	return dt == term.XSDDateTime || dt == term.XSDDate || dt == term.XSDTime
}

// parseDatePair attempts to parse two date/dateTime/time literals into time.Time values.
func parseDatePair(a, b term.Literal) (time.Time, time.Time, bool) {
	ta, okA := parseDateTime(a.Lexical(), a.Datatype())
	tb, okB := parseDateTime(b.Lexical(), b.Datatype())
	if okA && okB {
		return ta, tb, true
	}
	return time.Time{}, time.Time{}, false
}

func parseDateTime(s string, dt term.URIRef) (time.Time, bool) {
	var formats []string
	switch dt {
	case term.XSDDateTime:
		formats = []string{
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04:05.999999999Z07:00",
			"2006-01-02T15:04:05.999999999",
		}
	case term.XSDDate:
		formats = []string{
			"2006-01-02Z07:00",
			"2006-01-02",
		}
	case term.XSDTime:
		formats = []string{
			"15:04:05Z07:00",
			"15:04:05",
			"15:04:05.999999999Z07:00",
			"15:04:05.999999999",
		}
	default:
		return time.Time{}, false
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
