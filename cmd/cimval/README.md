# cimval

`cimval` is a command-line tool for validating CGMES (Common Grid Model Exchange Standard) XML files against SHACL rules and custom CIM constraints.

## Features

- **Profile Support**: Validate EQ, SSH, TP, SV, DL, DY, SC, GL, OP, and EQBD profiles.
- **SHACL Integration**: Runs over 4,000 generated checks derived from standard ENTSO-E SHACL files.
- **Rule Silencing**: Suppress known false positives or profile-specific limitations.
- **Rich Metadata**: Reports include rule names, detailed messages, and descriptions.
- **Flexible Output**: Supports both human-readable text and JSON formats.

## Usage

```bash
cimval [options] <xml-file1> [<xml-file2> ...]
```

### Options

| Flag | Description |
| :--- | :--- |
| `-profile` | Comma-separated list of profiles to check (e.g., `EQ,SSH,TP`). Default: all. |
| `-silence` | Comma-separated list of Rule IDs to ignore (e.g., `dl:DiagramObject.IdentifiedObject-valueType`). |
| `-json` | Output results in structured JSON format. |
| `-solved` | Enable SolvedMAS checks (default: false). |
| `-notsolved` | Enable NotSolvedMAS checks (default: true). |
| `-common` | Enable Common/AllProfiles rules (default: true). |

### Examples

**Standard validation of a full model:**
```bash
cimval -profile EQ,SSH,TP,SV,DL PST_Type1_*.xml
```

**Validating while silencing known cross-profile issues:**
```bash
cimval -silence dl:DiagramObject.IdentifiedObject-valueType,sv:SvStatus.ConductingEquipment-valueType *.xml
```

**Verifying standard ENTSO-E test configurations:**
```bash
cimval -profile EQ,SSH,TP,SV,DL \
  -silence dl:DiagramObject.IdentifiedObject-valueType,sv:SvStatus.ConductingEquipment-valueType \
  CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/*.xml
```

**Exporting violations to JSON for further processing:**
```bash
cimval -json data.xml > violations.json
```

## Exit Codes

- `0`: Validation successful (or only `sh:Warning`/`sh:Info` findings).
- `1`: Found one or more `sh:Violation` severity findings.
