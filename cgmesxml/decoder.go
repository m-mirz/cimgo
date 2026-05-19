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

type CIMProfile struct {
	ModelId          string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	ModelDependentOn *struct {
		MRID string `xml:"resource,attr"`
	} `xml:"Model.DependentOn,omitempty"`
	ModelCreated              string `xml:"Model.created"`
	ModelDescription          string `xml:"Model.description"`
	ModelModelingAuthoritySet string `xml:"Model.modelingAuthoritySet"`
	ModelProfile              string `xml:"Model.profile"`
	ModelScenarioTime         string `xml:"Model.scenarioTime"`
	ModelVersion              int    `xml:"Model.version"`
}

type CIMDataset struct {
	Profiles []*CIMProfile
	Elements cimstructs.CIMDataset
}

// DecodeProfiles decodes each reader concurrently into a separate CIMDataset,
// then merges them into cimData in input order. Callers control merge precedence
// by ordering the readers slice (earlier entries win on field conflicts).
func DecodeProfiles(readers []io.Reader, cimData *cimstructs.CIMDataset) (*cimstructs.CIMDataset, error) {
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
	for _, elem := range src.Elements {
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
			return cimData, nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			normalizeRDFAbout(&t)
			labelParts := strings.Split(t.Name.Local, ".")
			labelEnd := labelParts[len(labelParts)-1]

			if _, ok := cimstructs.StructMap[labelEnd]; ok {
				node := cimstructs.StructMap[labelEnd]()

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
