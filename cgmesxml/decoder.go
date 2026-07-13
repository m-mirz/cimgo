package cgmesxml

import (
	"cimgo/cimstructs"
	"cimgo/cimxml"
	"encoding/xml"
	"io"
	"strings"
	"sync"
)

const rdfNS = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"

// normalizeRDFAbout converts rdf:about="#UUID" to rdf:ID="UUID" so both
// identifier forms map to the same xml:"ID,attr" struct field on cimbase.Base.
func normalizeRDFAbout(t *xml.StartElement) {
	for i := range t.Attr {
		if t.Attr[i].Name.Local == "about" && t.Attr[i].Name.Space == rdfNS {
			t.Attr[i].Name.Local = "ID"
			t.Attr[i].Name.Space = ""
			t.Attr[i].Value = strings.TrimPrefix(t.Attr[i].Value, "#")
			return
		}
	}
}

// mridFromAttrs extracts the mRID from a start element's already-normalized
// attributes (see normalizeRDFAbout) — the same "ID" local-name lookup that
// cimbase.Base's xml:"ID,attr" field uses, computed ahead of decode so the
// decoder can attribute FieldOccurrences correctly without waiting for the
// decoded node.
func mridFromAttrs(attrs []xml.Attr) string {
	for _, a := range attrs {
		if a.Name.Local == "ID" {
			return a.Value
		}
	}
	return ""
}

// mergeOccurrences overwrites per-(mRID, tag) counts in dst with src's,
// matching cimbase.DeepMerge's "later file wins" policy for the scalar value
// itself. Not summed across files — cross-file duplicate detection is out of
// scope (see plan doc); only a single file's single parse of one element is
// ever meaningful for sh:maxCount purposes.
func mergeOccurrences(dst, src map[string]map[string]int) {
	for mrid, tags := range src {
		dstTags := dst[mrid]
		if dstTags == nil {
			dstTags = make(map[string]int, len(tags))
			dst[mrid] = dstTags
		}
		for tag, n := range tags {
			dstTags[tag] = n
		}
	}
}

// DecodeProfilesSeparate decodes each reader concurrently (one goroutine per
// reader) into its own CIMDataset, returned in the same order as readers,
// without merging. Callers that need both a merged view and per-file
// isolation (e.g. cmd/cimcli, which detects each file's profile and pulls
// EQBD BaseVoltage IDs from the isolated EQBD dataset before merging) should
// call this directly instead of DecodeProfiles, which discards the
// individual results after merging.
func DecodeProfilesSeparate(readers []io.Reader) ([]*cimstructs.CIMDataset, error) {
	results := make([]*cimstructs.CIMDataset, len(readers))
	errs := make([]error, len(readers))

	var wg sync.WaitGroup
	wg.Add(len(readers))
	for i, r := range readers {
		go func(i int, r io.Reader) {
			defer wg.Done()
			results[i], errs[i] = DecodeProfile(r, nil)
		}(i, r)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// DecodeProfiles decodes each reader concurrently into a separate CIMDataset,
// then merges them into cimData in input order. Callers control merge precedence
// by ordering the readers slice (earlier entries win on field conflicts).
func DecodeProfiles(readers []io.Reader, cimData *cimstructs.CIMDataset) (*cimstructs.CIMDataset, error) {
	results, err := DecodeProfilesSeparate(readers)
	if err != nil {
		return nil, err
	}

	if cimData == nil {
		cimData = cimstructs.NewCIMDataset()
	}
	for _, r := range results {
		if err := MergeInto(cimData, r); err != nil {
			return nil, err
		}
	}
	return cimData, nil
}

// MergeInto adds all elements from src into dst, merging any objects with
// matching mRIDs via DeepMerge.
func MergeInto(dst, src *cimstructs.CIMDataset) error {
	mergeOccurrences(dst.FieldOccurrences, src.FieldOccurrences)
	for _, elem := range src.ByID {
		if err := dst.AddElement(elem); err != nil {
			return err
		}
	}
	return nil
}

func DecodeProfile(r io.Reader, cimData *cimstructs.CIMDataset) (*cimstructs.CIMDataset, error) {
	if cimData == nil {
		cimData = cimstructs.NewCIMDataset()
	}
	dec := cimxml.NewDecoder(r)

	for {
		token, err := dec.Token()
		if err != nil && err != io.EOF {
			return cimData, err
		}

		if err == io.EOF {
			// slog.Debug("Reached end of file")
			mergeOccurrences(cimData.FieldOccurrences, dec.FieldOccurrences)
			return cimData, nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			normalizeRDFAbout(&t)
			labelParts := strings.Split(t.Name.Local, ".")
			labelEnd := labelParts[len(labelParts)-1]

			if _, ok := cimstructs.StructMap[labelEnd]; ok {
				node := cimstructs.StructMap[labelEnd]()

				dec.CurrentMRID = mridFromAttrs(t.Attr)
				if err := dec.DecodeElement(node, &t); err != nil {
					return cimData, err
				}

				if err := cimData.AddElement(node); err != nil {
					return cimData, err
				}
			}

		case xml.EndElement:
			//labelParts := strings.Split(t.Name.Local, ".")
			//labelEnd := labelParts[len(labelParts)-1]
			// slog.Debug("Found", "EndElement", labelEnd)
		case xml.CharData:
			if str := strings.TrimSpace(string(t)); len(str) > 0 {
				// slog.Debug("Found", "CharData", str)
			}
		case xml.Comment:
			// slog.Debug("Found", "Comment", string(t))
		case xml.Directive:
			// slog.Debug("Found", "Directive", string(t))
		case xml.ProcInst:
			// slog.Debug("Found", "ProcInst target", t.Target, "ProcInst inst", string(t.Inst))
		}
	}
}
