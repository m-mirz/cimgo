package validation

import (
	"cimgo/cimstructs"
	"sort"
)

var profURIToShortName = map[string]string{
	profEQ:   "EQ",
	profSSH:  "SSH",
	profTP:   "TP",
	profSV:   "SV",
	profDY:   "DY",
	profSC:   "SC",
	profDL:   "DL",
	profGL:   "GL",
	profOP:   "OP",
	profEQBD: "EQBD",
}

// DetectConfig inspects the dataset's model headers and returns a Config with
// Profiles, Solved, and NotSolved populated from what is actually present.
// Solved is true when an SV profile is found (power-flow results available).
// NotSolved is the complement of Solved.
// Common, Quality, SilencedRules, and EQBDBaseVoltageIDs are left at zero values.
func DetectConfig(dataset *cimstructs.CIMElementList) Config {
	seen := make(map[string]bool)
	collect := func(m *cimstructs.Model) {
		if short, ok := profURIToShortName[profileURI(m)]; ok {
			seen[short] = true
		}
	}
	for _, fm := range dataset.FullModels {
		collect(&fm.Model)
	}
	for _, dm := range dataset.DifferenceModels {
		collect(&dm.Model)
	}

	profiles := make([]string, 0, len(seen))
	for p := range seen {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)

	isSolved := seen["SV"]
	return Config{
		Profiles:  profiles,
		Solved:    isSolved,
		NotSolved: !isSolved,
	}
}
