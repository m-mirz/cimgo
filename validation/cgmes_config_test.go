package validation

import (
	"bytes"
	"cimgo/cimstructs"
	"cimgo/cgmesxml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePSTType1EQ(t *testing.T) {
	// Conformant data should produce zero sh:Violation findings for EQ.
	// However, the Simple profiles (600-2) in CGMES have some cross-profile
	// limitations that trigger violations in standard test data.
	// By silencing these known issues, we can ensure the rest of the model is valid.
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 0 violations, 1 PROF10-EQ sh:Info)", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh:Violation findings for PST Type 1 data after fixing known issues, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
}

func TestValidatePSTType2EQ(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type2/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 0 violations, 1 PROF10-EQ sh:Info)", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh:Violation findings for PST Type 2 data, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
}

func TestValidateSmallGridMerged(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/SmallGrid/SmallGrid-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "GL", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch {
			case v.RuleID == "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case v.Class == "EquivalentInjection":
				counts["equivalentInjection"]++
			case v.RuleID == "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			}
		}
	}
	t.Logf("Total: %d violations", counts["total"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["equivalentInjection"] != 4 {
		t.Errorf("EquivalentInjection range: expected 4, got %d", counts["equivalentInjection"])
	}
	if counts["syncMachine"] != 23 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 23, got %d", counts["syncMachine"])
	}
	if counts["total"] != 28 {
		t.Errorf("total violations: expected 28, got %d", counts["total"])
	}
}

func TestValidateMicroGridBaseCase(t *testing.T) {
	const path = "../CGMES-Test-Configurations/v3.0/MicroGrid/MicroGid-BaseCase/MicroGrid-BaseCase-Merged/"
	dataset := loadDirectory(t, path)

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "GL", "DY", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
		},
		EQBDBaseVoltageIDs: loadEQBDBaseVoltageIDs(t, path),
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch {
			case v.RuleID == "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case v.Class == "EquivalentInjection":
				counts["equivalentInjection"]++
			case strings.HasPrefix(v.RuleID, "coreeqc:TapChanger"):
				counts["tapChanger"]++
			case v.RuleID == "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case v.RuleID == "all600:All-GENC1":
				counts["genc1"]++
			}
		}
		if v.RuleID == "eqbd2:EQBD2" {
			counts["eqbd2"]++
		}
	}
	t.Logf("Total: %d violations, %d eqbd2 warnings", counts["total"], counts["eqbd2"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["equivalentInjection"] != 20 {
		t.Errorf("EquivalentInjection range: expected 20, got %d", counts["equivalentInjection"])
	}
	if counts["tapChanger"] != 5 {
		t.Errorf("TapChanger cardinality: expected 5, got %d", counts["tapChanger"])
	}
	if counts["syncMachine"] != 4 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 4, got %d", counts["syncMachine"])
	}
	if counts["genc1"] != 1 {
		t.Errorf("All-GENC1: expected 1, got %d", counts["genc1"])
	}
	if counts["total"] != 31 {
		t.Errorf("total violations: expected 31, got %d", counts["total"])
	}
	if counts["eqbd2"] != 4 {
		t.Errorf("EQBD2 (BaseVoltage not in boundary): expected 4, got %d", counts["eqbd2"])
	}
}

func TestValidateMicroGridType1(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/MicroGrid/MicroGrid-Type1/MicroGrid-Type1-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "GL", "DY", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch {
			case v.RuleID == "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case v.Class == "EquivalentInjection":
				counts["equivalentInjection"]++
			case strings.HasPrefix(v.RuleID, "coreeqc:TapChanger"):
				counts["tapChanger"]++
			case v.RuleID == "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case v.RuleID == "all600:All-GENC1":
				counts["genc1"]++
			}
		}
	}
	t.Logf("Total: %d violations", counts["total"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["equivalentInjection"] != 20 {
		t.Errorf("EquivalentInjection range: expected 20, got %d", counts["equivalentInjection"])
	}
	if counts["tapChanger"] != 5 {
		t.Errorf("TapChanger cardinality: expected 5, got %d", counts["tapChanger"])
	}
	if counts["syncMachine"] != 4 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 4, got %d", counts["syncMachine"])
	}
	if counts["genc1"] != 1 {
		t.Errorf("All-GENC1: expected 1, got %d", counts["genc1"])
	}
	if counts["total"] != 31 {
		t.Errorf("total violations: expected 31, got %d", counts["total"])
	}
}

func TestValidateMicroGridType2(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/MicroGrid/MicroGrid-Type2/MicroGrid-Type2-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "GL", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch {
			case v.RuleID == "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case v.Class == "EquivalentInjection":
				counts["equivalentInjection"]++
			case strings.HasPrefix(v.RuleID, "coreeqc:TapChanger"):
				counts["tapChanger"]++
			case v.RuleID == "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case v.RuleID == "all600:All-GENC1":
				counts["genc1"]++
			}
		}
	}
	t.Logf("Total: %d violations", counts["total"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["equivalentInjection"] != 24 {
		t.Errorf("EquivalentInjection range: expected 24, got %d", counts["equivalentInjection"])
	}
	if counts["tapChanger"] != 8 {
		t.Errorf("TapChanger cardinality: expected 8, got %d", counts["tapChanger"])
	}
	if counts["syncMachine"] != 4 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 4, got %d", counts["syncMachine"])
	}
	if counts["genc1"] != 1 {
		t.Errorf("All-GENC1: expected 1, got %d", counts["genc1"])
	}
	if counts["total"] != 38 {
		t.Errorf("total violations: expected 38, got %d", counts["total"])
	}
}

func TestValidateSvedalaMerged(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/Svedala/Svedala-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch v.RuleID {
			case "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case "prof10:PROF10":
				counts["prof10"]++
			}
		}
	}
	t.Logf("Total: %d violations", counts["total"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["syncMachine"] != 38 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 38, got %d", counts["syncMachine"])
	}
	if counts["prof10"] != 0 {
		t.Errorf("prof10:PROF10: expected 0, got %d", counts["prof10"])
	}
	if counts["total"] != 39 {
		t.Errorf("total violations: expected 39, got %d", counts["total"])
	}
}

func TestValidateFullGridMerged(t *testing.T) {
	const path = "../CGMES-Test-Configurations/v3.0/FullGrid/FullGrid-Merged/"
	dataset := loadDirectory(t, path)

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "SC", "OP", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
		EQBDBaseVoltageIDs: loadEQBDBaseVoltageIDs(t, path),
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch {
			case v.RuleID == "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case strings.HasPrefix(v.RuleID, "coreeqc:TapChanger"):
				counts["tapChanger"]++
			case v.RuleID == "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case v.RuleID == "sc:PowerTransformerEnd.phaseAngleClock-cardinality":
				counts["phaseAngleClock"]++
			case strings.HasPrefix(v.RuleID, "op:"):
				counts["operation"]++
			case strings.HasSuffix(v.RuleID, "-datatype"):
				counts["datatype"]++
			}
		}
		if v.RuleID == "eqbd2:EQBD2" {
			counts["eqbd2"]++
		}
	}
	t.Logf("Total: %d violations, %d eqbd2 warnings", counts["total"], counts["eqbd2"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["tapChanger"] != 15 {
		t.Errorf("TapChanger cardinality: expected 15, got %d", counts["tapChanger"])
	}
	if counts["syncMachine"] != 6 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 6, got %d", counts["syncMachine"])
	}
	if counts["phaseAngleClock"] != 23 {
		t.Errorf("PowerTransformerEnd.phaseAngleClock-cardinality: expected 23, got %d", counts["phaseAngleClock"])
	}
	if counts["operation"] != 12 {
		t.Errorf("OP profile violations: expected 12, got %d", counts["operation"])
	}
	if counts["datatype"] != 10 {
		t.Errorf("schedule datatype violations: expected 10, got %d", counts["datatype"])
	}
	if counts["total"] != 87 {
		t.Errorf("total violations: expected 87, got %d", counts["total"])
	}
	if counts["eqbd2"] != 5 {
		t.Errorf("EQBD2 (BaseVoltage not in boundary): expected 5, got %d", counts["eqbd2"])
	}
}

func TestValidateRealGridMerged(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV"},
		Common:   true,
		SilencedRules: []string{
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch v.RuleID {
			case "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case "coreeqc:TapChanger.neutralStep-cardinality":
				counts["tapChangerNeutral"]++
			case "coreeqc:TapChanger.normalStep-cardinality":
				counts["tapChangerNormal"]++
			}
		}
		if v.RuleID == "prof10:PROF10" && v.Severity == "sh:Info" {
			counts["prof10Info"]++
		}
	}
	t.Logf("Total: %d violations, %d prof10 info", counts["total"], counts["prof10Info"])

	if counts["syncMachine"] != 1346 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 1346, got %d", counts["syncMachine"])
	}
	if counts["tapChangerNeutral"] != 2 {
		t.Errorf("TapChanger.neutralStep-cardinality: expected 2, got %d", counts["tapChangerNeutral"])
	}
	if counts["tapChangerNormal"] != 2 {
		t.Errorf("TapChanger.normalStep-cardinality: expected 2, got %d", counts["tapChangerNormal"])
	}
	if counts["total"] != 1350 {
		t.Errorf("total violations: expected 1350, got %d", counts["total"])
	}
	if counts["prof10Info"] != 1 {
		t.Errorf("PROF10-EQ sh:Info (missing EQBD ref): expected 1, got %d", counts["prof10Info"])
	}
}

func TestValidateMiniGridMerged(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/MiniGrid/MiniGrid-Merged/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "EQBD"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
			"eqbd:Terminal.ConductingEquipment-valueType",
			"sv456cpi:SvSwitch.Switch-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	counts := map[string]int{}
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			counts["total"]++
			switch v.RuleID {
			case "eq600-2:GeographicalRegion-EQ__4":
				counts["geographicalRegion"]++
			case "ssh:SynchronousMachine.referencePriority-cardinality":
				counts["syncMachine"]++
			case "ssh:ExternalNetworkInjection.referencePriority-cardinality":
				counts["externalNetworkInjection"]++
			}
		}
	}
	t.Logf("Total: %d violations", counts["total"])

	if counts["geographicalRegion"] != 1 {
		t.Errorf("GeographicalRegion: expected 1, got %d", counts["geographicalRegion"])
	}
	if counts["syncMachine"] != 2 {
		t.Errorf("SynchronousMachine.referencePriority-cardinality: expected 2, got %d", counts["syncMachine"])
	}
	if counts["externalNetworkInjection"] != 2 {
		t.Errorf("ExternalNetworkInjection.referencePriority-cardinality: expected 2, got %d", counts["externalNetworkInjection"])
	}
	if counts["total"] != 5 {
		t.Errorf("total violations: expected 5, got %d", counts["total"])
	}
}

func TestValidatePowerFlow(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PowerFlow/PowerFlow/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV"},
		Common:   true,
		SilencedRules: []string{
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	var errCount, infoEQBDCount, warnSubstationCount int
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
		if v.RuleID == "eq600:Substation-count" {
			warnSubstationCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh:Violation findings for PowerFlow data, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
	if warnSubstationCount != 1 {
		t.Errorf("expected 1 eq600:Substation-count sh:Warning, got %d", warnSubstationCount)
	}
}

func TestValidatePSTType3(t *testing.T) {
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerTable_Type3/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 0 violations, 1 PROF10-EQ sh:Info)", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh:Violation findings for PST Type 3 data, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
}

// readRealGridReaders reads all XML files from the RealGrid-Merged directory and
// returns them as byte slices (to allow repeated benchmark iterations without re-reading disk).
func readRealGridFiles(b *testing.B) [][]byte {
	b.Helper()
	const dir = "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/"
	entries, err := os.ReadDir(dir)
	if err != nil {
		b.Fatalf("failed to read RealGrid directory: %v", err)
	}
	var blobs [][]byte
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".xml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			b.Fatalf("failed to read %s: %v", e.Name(), err)
		}
		blobs = append(blobs, data)
	}
	return blobs
}

// BenchmarkRealGridValidation measures RunValidation on RealGrid (~115 MB, 4 files) with
// all profile validators running in parallel. Dataset loading is excluded from the timer.
func BenchmarkRealGridValidation(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV"},
		Common:   true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = RunValidation(dataset, cfg)
	}
}

// BenchmarkSmallGridValidation measures RunValidation on SmallGrid (7 profiles, ~14 MB),
// which has more parallelism headroom than RealGrid and is closer to a typical dataset.
func BenchmarkSmallGridValidation(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/SmallGrid/SmallGrid-Merged/")
	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL", "GL", "EQBD"},
		Common:   true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = RunValidation(dataset, cfg)
	}
}

// BenchmarkRealGridLoadSequential measures loading RealGrid (~115 MB, 4 files) with
// the original sequential DecodeProfile approach.
func BenchmarkRealGridLoadSequential(b *testing.B) {
	blobs := readRealGridFiles(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		dataset := cimstructs.NewCIMElementList()
		for _, blob := range blobs {
			if _, err := cgmesxml.DecodeProfile(bytes.NewReader(blob), dataset); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkRealGridLoadParallel measures loading RealGrid (~115 MB, 4 files) with
// the new parallel DecodeProfiles approach.
func BenchmarkRealGridLoadParallel(b *testing.B) {
	blobs := readRealGridFiles(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		readers := make([]io.Reader, len(blobs))
		for j, blob := range blobs {
			readers[j] = bytes.NewReader(blob)
		}
		if _, err := cgmesxml.DecodeProfiles(readers, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealGridValidateEQ(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	b.ResetTimer()
	for b.Loop() { _ = ValidateEQProfile(dataset) }
}
func BenchmarkRealGridValidateSSH(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	b.ResetTimer()
	for b.Loop() { _ = ValidateSSHProfile(dataset) }
}
func BenchmarkRealGridValidateTP(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	b.ResetTimer()
	for b.Loop() { _ = ValidateTPProfile(dataset) }
}
func BenchmarkRealGridValidateSV(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	b.ResetTimer()
	for b.Loop() { _ = ValidateSVProfile(dataset) }
}
func BenchmarkRealGridValidateCommon(b *testing.B) {
	dataset := loadDirectory(b, "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged/")
	b.ResetTimer()
	for b.Loop() { _ = ValidateCommonRulesSPARQL(dataset) }
}
