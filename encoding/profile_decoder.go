package encoding

import (
	"cimgo/cimgostructs"
	"cimgo/cimxml"
	"encoding/xml"
	"io"
	"strings"
)

type CIMProfile struct {
	ModelId          string `xml:"rdf:about,attr"`
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
	Elements cimgostructs.CIMElementList
}

func DecodeProfile(r io.Reader, cimData *cimgostructs.CIMElementList) (*cimgostructs.CIMElementList, error) {
	if cimData == nil {
		cimData = cimgostructs.NewCIMElementList()
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
			labelParts := strings.Split(t.Name.Local, ".")
			labelEnd := labelParts[len(labelParts)-1]

			if _, ok := cimgostructs.StructMap[labelEnd]; ok {
				node := cimgostructs.StructMap[labelEnd]()

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
