# CIMgo

CIMgo is a Go library and CLI for working with CGMES/CIM electrical grid data. It parses CGMES XML instance files, validates them against SHACL and SPARQL based rules, and generates typed Go structs and Protobuf definitions from CIM RDF schemas.

## cimcli — CIM CGMES CLI

`cimcli` is a command-line tool for working with CGMES (Common Grid Model Exchange Standard) XML files.

### Download

Pre-built binaries are attached to each [GitHub release](https://github.com/m-mirz/cimgo/releases/latest):

| Platform | File |
|----------|------|
| Linux (x86-64) | `cimcli-linux-amd64` |
| Windows (x86-64) | `cimcli-windows-amd64.exe` |
| Mac (arm64) | `cimcli-darwin-arm64` |
| Mac (x86-64) | `cimcli-darwin-amd64` |

**Linux** — make the binary executable after downloading:

```bash
chmod +x cimcli-linux-amd64
```

**Windows** — the `.exe` can be run directly from PowerShell or CMD.

### Commands

#### validate

Validates CGMES XML instance files against CGMES SHACL and SPARQL rules. Profiles, solved/not-solved state, and EQBD base voltage IDs are detected automatically from the file headers.

- Runs over 4,000 generated checks derived from standard ENTSO-E SHACL files.
- Supports EQ, SSH, TP, SV, DL, DY, SC, GL, OP, and EQBD profiles.
- Rule silencing: `dl:DiagramObject.IdentifiedObject-valueType` and `sv:SvStatus.ConductingEquipment-valueType` are silenced by default; additional rules can be suppressed via `-silence`.
- Human-readable text and JSON output formats.

```bash
cimcli validate [options] <xml-file1> [<xml-file2> ...]
```

| Flag | Description |
| :--- | :--- |
| `-profile` | Comma-separated list of profiles to check (e.g., `EQ,SSH,TP`). Default: auto-detected. |
| `-silence` | Comma-separated list of additional Rule IDs to ignore. |
| `-json` | Output violations in structured JSON format. |
| `-solved` | Enable SolvedMAS checks (default: auto-detected). |
| `-notsolved` | Enable NotSolvedMAS checks (default: auto-detected). |
| `-common` | Enable Common/AllProfiles rules (default: true). |
| `-quality` | Enable CIMdesk-style modeling quality checks (default: false). |

Exit code is `0` when no `sh:Violation`-severity findings are present, `1` otherwise.

#### convert

Converts between CGMES XML, JSON, and binary Protobuf. Input format is
auto-detected from the file extension (`.json` vs `.xml`).

```bash
cimcli convert [options] <file1> [<file2> ...]
```

| Flag | Description |
| :--- | :--- |
| `-to` | Output format: `json` (default), `proto`, or `xml`. |
| `-out` | Output file for `json` (default: stdout) or `proto` (default: `output.pb`), or output directory for `xml` (default: current directory). |
| `-profile` | Comma-separated profile codes for `-to xml` (e.g. `EQ,SSH,TP`). Default: all profiles. |

**XML → JSON** — merges one or more CGMES XML files and writes the combined
dataset as JSON to stdout. Each element carries a `_type` field with its CIM
class name, enabling the output to be converted back to XML.

**XML → Protobuf** — converts the merged dataset to a binary `CIMElementList`
proto message.

**JSON → XML** — reads a JSON file produced by `-to json` and writes one
CGMES XML file per profile code into the output directory.

### Examples

**Validate dataset:**
```bash
# Linux
./cimcli-linux-amd64 validate EQ.xml SSH.xml TP.xml SV.xml

# Windows
cimcli-windows-amd64.exe validate EQ.xml SSH.xml TP.xml SV.xml
```

**Validate dataset against specific profile rules:**
```bash
./cimcli-linux-amd64 validate -profile EQ,SSH,TP,SV,DL PST_Type1_*.xml
```

**Validate ENTSO-E test configurations:**
```bash
./cimcli-linux-amd64 validate -profile EQ,SSH,TP,SV,DL \
  CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/*.xml
```

**Output violations as JSON:**
```bash
./cimcli-linux-amd64 validate -json -profile EQ,SSH,TP,SV,DL \
  CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/*.xml
```

**Convert XML files to JSON (stdout):**
```bash
./cimcli-linux-amd64 convert EQ.xml SSH.xml
```

**Convert XML files to JSON (file):**
```bash
./cimcli-linux-amd64 convert -out dataset.json EQ.xml SSH.xml
```

**Convert XML files to binary Protobuf:**
```bash
./cimcli-linux-amd64 convert -to proto -out dataset.pb EQ.xml SSH.xml
```

**Convert JSON back to CGMES XML profiles:**
```bash
./cimcli-linux-amd64 convert -to xml -out ./output/ dataset.json
```

**Convert only specific profiles back to XML:**
```bash
./cimcli-linux-amd64 convert -to xml -profile EQ,SSH -out ./output/ dataset.json
```

---

## How to Build and Run

### Setup

Make sure that you have cloned the repo recursively to include the CGMES schema files from ENTSO-E

    git clone --recurse-submodules [...]

or clone the submodule in a second step

    git submodule update --init --recursive

Ensure that GOPATH is set and included in your PATH.

For the protobuf code generation, you also require the proto compiler

    sudo apt-get install -y protobuf-compiler
    # install tools from mod file
    go get tool

### Commands

```bash
# Generate all code (must run before build after schema changes)
go generate ./...

# Build
go build -v ./...

# Run all tests
go test -v ./...

# Run a single test
go test -v ./path/to/package -run TestName
```

## Profiling validation performance

Benchmarks covering end-to-end validation and per-profile breakdown live in
`validation/cgmes_config_test.go`.

**Rank profiles by wall time** (quick, no pprof overhead):

```bash
go test -run='^$' \
    -bench='BenchmarkRealGridValidate(EQ|SSH|TP|SV|Common)$' \
    -benchtime=3x ./validation/
```

**Full pipeline with CPU and memory profiles** (RealGrid, ~115 MB, 4 profiles):

```bash
go test -run='^$' \
    -bench=BenchmarkRealGridValidation \
    -benchtime=3x \
    -cpuprofile=cpu.prof \
    -memprofile=mem.prof \
    ./validation/
```

**Inspect results:**

```bash
go tool pprof -http=:6060 cpu.prof   # flame graph + top + source view
go tool pprof -http=:6061 mem.prof   # set Sample dropdown to alloc_space
```


## Architecture

The code generation process follows these main steps:

1.  **Schema Loading:** The tool begins by finding and parsing the relevant CIM RDF schema or TTL SHACL files based on the specified version and profiles.
2.  **Schema Processing:** It processes the parsed RDF or SHACL data into an internal, language-agnostic representation of CIM classes, properties, datatypes, their relationships and rules.
3.  **Code Generation:** Using Go's `text/template` engine, it feeds the internal representation into language-specific templates (`lang-templates/*.tmpl`) to generate the final source code files.

## SHACL Validation

`cmd/shaclgen` translates each constraint in the CGMES SHACL Turtle files into 
a Go `Check<...>` function under `shaclgen/`. The 205 SPARQL constraints (all 
currently implemented, see below) are not generated by `cmd/shaclgen` but 
instead implemented as hand-written Go functions.

Across 74 TTL SHACL files there are ~10,000 constraints total. About half of these generate
code and the other half is skipped: The skipped ones are either structurally satisfied 
by the Go type system (generating a check would never fire or would produce false positives)
or cannot be conducted due to upstream SHACL TTL defects. Counts of generated-vs-skipped by
CGMES profile group are in ["Generated SHACL Rules by Profile"](#generated-shacl-rules-by-profile) below.

### Simplification rules applied during import

Before code generation, `shaclimport.SimplifyFileResults` normalises each
property shape's constraint list. Rules are applied in order; a constraint
that matches a rule is either dropped or rewritten and is not passed on to
`cmd/shaclgen`.

*The simplifications applied here do not necessarily work for other validation
engines and are specific to this tool.*

| Rule | Constraint removed / rewritten | Reason |
|------|-------------------------------|--------|
| 1 | `sh:nodeKind sh:Literal` when any `sh:datatype` is also present | `sh:datatype` already implies the value is a literal; the `sh:nodeKind` check is redundant. |
| 2 | `sh:nodeKind sh:BlankNodeOrIRI` and `sh:nodeKind sh:IRI` unconditionally | Every IRI-typed CIM property is generated as a Go reference field (`*struct{ MRID string }`); the type system already enforces the IRI shape. |
| 3 | `sh:minCount 0` | Vacuously true — zero or more values are always acceptable. |
| 4 | `sh:datatype xsd:T` for native Go scalar types | The Go struct field is already typed (`int`, `float64`, `bool`, `string`), so the XML decoder rejects malformed input before validation. Dropped for: all integer variants, `float`/`double`/`decimal`, `boolean`, `string`/`normalizedString`/`token`. Non-native types (`dateTime`, `gMonthDay`, `anyURI`) are **not** dropped. |
| 5 | `sh:in` with a single value → rewritten as `sh:hasValue` | A one-element allow-list is semantically identical to an exact-value check. |
| 6 | `sh:minCount 0` + `sh:maxCount 1` → synthetic `sh:Optional` | The pair means "0 or 1 values". `sh:minCount 0` is dropped by Rule 3; the matching `sh:maxCount 1` is replaced by a single `Optional` sentinel to record the upper bound without implying a presence requirement. |
| 7 | `sh:minCount 1` + `sh:maxCount 1` → synthetic `sh:Required` | The pair means "exactly 1 value". Both constraints are collapsed into a single `Required` sentinel, avoiding duplicate presence checks. |

### Skipped constraints

The two tables below summarise every skipped-constraint category and count as reported by `go run ./cmd/shaclgen -skip-report`, split by whether the skip represents work left to do; their totals sum to 9998, matching the ["Generated SHACL Rules by Profile"](#generated-shacl-rules-by-profile) table's `Skipped` column below exactly. The two `sh:nodeKind`/`sh:datatype` rows are the constraints dropped by Rules 1, 2, and 4 above. The SPARQL row is not a type-system tautology — those constraints are implemented separately as hand-written Go functions (see ["SPARQL Check Coverage"](#sparql-check-coverage) below). The `Upstream SHACL TTL defects` row is individually broken out after the table.

Row labels below are kept in sync with cimoxide's equivalent table where the underlying reason
is the same — see ["Comparing skip categories with cimoxide"](#comparing-skip-categories-with-cimoxide)
right after the table for how the two tools' categories do (and don't) line up.

#### Does not require fix

These are either handled by an alternative method (hand-written functions), structurally
guaranteed by the Go type system or the decoded representation so a generated check could
never fire, or defects in the upstream ENTSO-E TTL files that only ENTSO-E can fix.

| Count | Constraint | Reason |
|------:|-----------|--------|
| 3472 | `sh:nodeKind` simplified | Structurally satisfied by the Go type system (Rules 1–2 above). |
| 3126 | `sh:datatype` simplified | Structurally satisfied by native Go scalar types (Rule 4 above). |
| 182 | SPARQL-derived constraints (`sh:sparql`, plus SPARQL-based `sh:target`) | Not generated by `shaclgen`; implemented as hand-written Go functions instead (see "SPARQL Check Coverage" below). Includes the four `C:600:ALL:NA:PROF10` sub-constraints whose shape uses a SPARQL-defined `sh:target` (see the `targetSubjectsOf`/`targetObjectsOf` row below for the sibling case) — `PROF10` itself is already implemented as a hand-written check (see "C:600 conformance rules" below). This total isn't directly comparable to the "SPARQL Check Coverage" table's TTL Total — see the note in `cmd/shaclgen/classify.go`. |
| 28 | `sh:maxCount 1` on multi-hop paths | Every hop in a CIM reference path is a 0..1 pointer, so the count is always ≤ 1. Cross-tool: matches cimoxide's identically-named row (28) exactly. |
| 13 | Upstream SHACL TTL defects | Cannot be conducted because the constraint itself references a misspelled or non-existent field/class in the upstream CGMES SHACL files (see "Upstream SHACL TTL defects" below for the individual cases). |
| 7 | `sh:nodeKind` on `rdf:type` paths | The Go struct type is fixed at decode time, so the RDF type is always correct. |
| 5 | `sh:class` vacuously true (inverse-index already type-asserts) | These are `sh:class` sub-constraints on property shapes that also carry `sh:minCount`/`sh:maxCount` (e.g. `C:301:EQ:Switch:numberOfTerminals`). The count sub-constraint is generated normally. The `sh:class` sub-constraint is skipped because the inverse index prelude already type-asserts each scanned object to the referrer class — every object that reaches the check has already passed that assertion, so the class constraint is tautological. Cross-tool: same reason as cimoxide's identically-named row (6) — the count differs only by dedup granularity. |
| 5 | Unsupported `sh:target` kind (`targetSubjectsOf`/`targetObjectsOf`) | Five property shapes (`C:600:ALL:NA:FBOD4` and four `IdentifiedObject.*:stringLength` rules) use `sh:targetSubjectsOf`/`sh:targetObjectsOf` rather than `sh:targetClass`, so `shaclgen` has no concrete class to generate a check against — this holds regardless of the underlying constraint's own component. All 5 are `sh:SPARQLConstraintComponent` shapes too, and all 5 already have hand-written coverage in `validation/sparql_common.go`/`validation/sparql_common_solvedmas.go` under those exact `sh:name`s (see the 100%-coverage "SPARQL Check Coverage" table below) — so despite the skip, nothing here is actually unchecked. Cross-tool: matches cimoxide's identically-named row (5) exactly, including the same 18 distinct `sh:name`s — see ["Comparing skip categories with cimoxide"](#comparing-skip-categories-with-cimoxide) below. |
| 3 | Cross-class `sh:lessThan` on sibling subtypes | The SHACL property shape reuses the same comparison across multiple target classes. The compared field (`xDirectSubtrans`, `xQuadSubtrans`, `xpp`) exists only on a sibling subtype (`SynchronousMachineTimeConstantReactance` or `AsynchronousMachineTimeConstantReactance`). These subtypes are mutually exclusive in valid CGMES data — a machine is either `SynchronousMachineTimeConstantReactance` or `SynchronousMachineEquivalentCircuit`, never both — so both operands can never be visible at the same time. The same-class cases (`SynchronousMachineTimeConstantReactance.statorLeakageReactance < xDirectSubtrans`, etc.) are generated normally. Cross-tool: part of the "Multi-hop / multi-segment path constraints" group, see below. |
| 3 | `sh:hasValue rdf:type rdf:Statements` on `forwardDifferences`/`reverseDifferences`/`preconditions` | The CGMES difference model format mandates that all elements in these collections are `rdf:Statement` resources. The referenced objects are not decoded into `CIMElementList` (they are RDF graph metadata, not CIM elements), so the constraint cannot be checked at runtime — but it cannot be violated by any well-formed CGMES difference model file. |
| 3 | Multi-segment `sh:required` on `rdf:Statements.subject/predicate/object` | The RDF specification mandates that every `rdf:Statement` resource has `subject`, `predicate`, and `object` predicates. Any `rdf:Statement` instance loaded from a valid RDF document already satisfies these constraints by definition. Cross-tool: part of the "Multi-hop / multi-segment path constraints" group, see below. |
| **6847** | **Total** | |

#### Might require fix

| Count | Constraint | Reason | Fix |
|------:|-----------|--------|-----|
| 2702 | `sh:required` on `float` fields | Float fields use `omitempty`; `0.0` is indistinguishable from absent after XML decode. A legitimately-zero physical quantity (e.g. `bch=0`, `r=0`, `b=0`) would always trigger a false positive. Range constraints (`sh:minExclusive`, `sh:minInclusive`) cover the must-be-positive subset where zero is genuinely invalid. **No cimoxide equivalent** — cimoxide's generated fields are `Option<f64>`, so presence is distinguishable from a genuine zero and it generates this check instead of skipping it. | Switch all float fields to `*float64`. |
| 223 | `sh:maxCount 1` on scalar fields | Duplicate XML elements silently overwrite the field; no violation is emitted. **No cimoxide equivalent** — cimoxide's decoder tracks per-field duplicate XML occurrences and generates this check instead of skipping it. | Detect multiple occurrences in the XML parser, or use a two-pass approach that parses into hashmaps first. |
| 115 | `sh:required` on `bool` fields | Bool fields use `omitempty`; `false` is indistinguishable from absent after XML decode. **No cimoxide equivalent** — cimoxide's generated fields are `Option<bool>`, so it generates this check instead of skipping it. | Switch all bool fields to `*bool`. |
| 111 | `sh:maxCount 1` on pointer fields | Pointer fields are either nil (0 values) or non-nil (1 value). **No cimoxide equivalent** — cimoxide's decoder tracks per-field duplicate XML occurrences and generates this check instead of skipping it. | Detect multiple occurrences in the XML parser, or use a two-pass approach that parses into hashmaps first. |
| **3151** | **Total** | | |

#### Comparing skip categories with cimoxide

cimgo's and cimoxide's skip-category tables cover the same 74 TTL files and the same 12,270
non-SPARQL constraints. Rows describing the identical underlying SHACL feature are given
matching wording and cross-referenced inline above:

- `sh:class` vacuously true (5 here vs. 6 in cimoxide — dedup differs: cimgo dedups per
  concrete target class, cimoxide per shape).
- "`sh:maxCount 1` on multi-hop paths" (28) and "Multi-segment `sh:required` on
  `rdf:Statements`" (3) match cimoxide's identically-named rows exactly; cimgo's "Cross-class
  `sh:lessThan` on sibling subtypes" (3) equivalents sit in cimoxide's "Attribute not found"
  row instead.
- The SPARQL-derived-constraints row (182) and `targetSubjectsOf`/`targetObjectsOf` (5) match
  cimoxide's identically-named rows exactly, including the same underlying `sh:name` sets
  (181/18 distinct names).
- `sh:required` on `float`/`bool` fields (2702 + 115) and `sh:maxCount 1` on scalar/pointer
  fields (223 + 111) have no cimoxide row at all, because cimoxide's generated fields are
  `Option<f64>`/`Option<bool>` and its decoder tracks per-field duplicate XML occurrences —
  both let cimoxide generate real checks instead of skipping them. This is exactly the "Fix"
  already listed for these rows above. There is no cimoxide-only row with no cimgo
  equivalent.

#### Upstream SHACL TTL defects

**Field name typos** — `sh:lessThan` references a misspelled field name; all four are in `61970-302_Dynamics-AP-Con-Complex-SHACL.ttl`:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `C:302:DY:GovHydroIEEE0.pmin:valueRangePair` | `sh:lessThan cim:GovHydroIEEE.pmax` — class suffix `0` missing; should be `GovHydroIEEE0.pmax` |
| `C:302:DY:PFVArType1IEEEVArController.vvtmin:valueRangePair` | `sh:lessThan cim:PVFArType1IEEEVArController.vvtmax` — prefix transposed; should be `PFVArType1IEEEVArController.vvtmax` |
| `C:302:DY:ExcDC1A.efdmin:valueRangePair` | `sh:lessThan cim:ExcDC1A.edfmax` — letters transposed; should be `efdmax` |
| `C:302:DY:PssIEEE4B.vhmin:valueRangePair` | `sh:lessThan cim:PssIEEE4V.vhmax` — class suffix wrong; should be `PssIEEE4B.vhmax` |

**Class name typo** — one property shape in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` references a misspelled class in its inverse path (reported as one entry covering all 4 concrete target classes):

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `SynchronousMachineDynamics.CrossCompoundTurbineGovernorDyanmics-cardinality` | `sh:inversePath cim:CrossCompoundTurbineGovernorDyanmics.SynchronousMachineDynamics` — "Dynamics" misspelled as "Dyanmics"; applied to `SynchronousMachineEquivalentCircuit`, `SynchronousMachineSimplified`, `SynchronousMachineTimeConstantReactance`, `SynchronousMachineUserDefined` |

**Class name capitalisation mismatch** — two SHACL files (`61970-457_Dynamics-AP-Con-Complex-Explicit-CrossProfile-SHACL.ttl` and `61970-457_Dynamics-AP-Con-Complex-Implicit-CrossProfile-SHACL.ttl`) reference `cim:CSConverter` (capital S), but all RDFS schema files consistently define the class as `cim:CsConverter`. The two defective shapes generate 2 skip instances:

| Shape | Defect |
|-------|--------|
| `dy457cpe:CSCDynamics.CSConverter-valueType` | `sh:in (cim:CSConverter)` — should be `cim:CsConverter` |
| `dy457cpi:CSCDynamics.CSConverter-valueType` | same defect in the implicit cross-profile file |

**Wrong field names in inverse paths** — four property shapes in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` reference field names that do not match the RDFS schema:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `SynchronousMachineDynamics.CrossCompoundTurbineGovernorDynamics-cardinality` | `sh:inversePath cim:CrossCompoundTurbineGovernorDynamics.SynchronousMachineDynamics` — no such field; RDFS defines `HighPressureSynchronousMachineDynamics` and `LowPressureSynchronousMachineDynamics`; applies to 4 concrete target classes |
| `CsConverter.CSCDynamics-cardinality` | `sh:inversePath cim:CSCDynamics.CsConverter` — capitalisation wrong; RDFS defines `CSCDynamics.CSConverter` |
| `VCompIEEEType2.GenICompensationForGenJ-cardinality` | `sh:inversePath cim:GenICompensationForGenJ.VCompIEEEType2` — capitalisation wrong; RDFS defines `GenICompensationForGenJ.VcompIEEEType2` |
| `WindContQIEC.WindTurbineType3or4IEC-cardinality` | `sh:inversePath cim:WindTurbineType3or4IEC.WindContQIEC` — capitalisation wrong; RDFS defines `WindTurbineType3or4IEC.WIndContQIEC` |

**Stale field reference** — one property shape in `61970-301_Operation-AP-Con-Complex-SHACL.ttl` references a field that was present in CGMES 2.4 but removed from the current CGMES 3.0 schema:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `C:301:OP:AccumulatorValue.value:valueRange` | `sh:minExclusive` on `cim:AccumulatorValue.value` — field removed in CGMES 3.0 |

**Non-existent target class** — one property shape in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` lists `cim:GovHydroIEEE1` in its `sh:targetClass` alongside several real classes. No such class exists in the CIM standard or in `cimstructs`; `shaclgen` silently skips it when resolving concrete target classes.

**Empty `sh:in` list** — one property shape in `61970-600-2_Operation-AP-Con-Simple-SHACL.ttl` has `sh:in ()` (reported as one entry covering all 4 concrete target classes):

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `Measurement.Terminal-valueType` | `sh:in ()` — empty allow-list; applied to `Accumulator`, `Analog`, `Discrete`, `StringMeasurement` |

### Generated SHACL Rules by Profile

The counts in the "Skipped constraints" table above are global totals across all
74 TTL files. `go run ./cmd/shaclgen -rule-report` also breaks the generated-vs-skipped split
down by CGMES profile group, using `ttlGroupLabel` in `cmd/shaclgen/ttl_report.go` — the same
classifier the SPARQL Check Coverage table below uses, so both tables' rows line up 1:1 (and
with cimoxide's equivalent tables): `Common / AllProfiles` absorbs `C:600 conformance` plus
anything without its own hand-written profile group (`AllProfiles`/`IdentifiedObjectCommon`/
`GeographicalLocation`/the plain Header file), and `Topology`/`DiagramLayout`/`Operation` each
get their own row. "Generated" counts distinct `(path, component, name)` rule patterns actually
code-generated for that group — *not* distinct `Check<...>` functions: a rule pattern applied
to N concrete classes generates N functions in `shaclgen/`, but counts as 1 here
(`uniqueCheckPatterns` in `cmd/shaclgen/specs.go`), matching how skip entries were already
deduped the same way. "Skipped" is every constraint dropped for one of the reasons above;
"Total" is their sum — i.e. the number of non-SPARQL SHACL constraints CGMES defines for that
profile group, independent of either tool's codegen capability.

This per-group `Total` (though not the `Generated`/`Skipped` split — see below) matches
cimoxide's equivalent table exactly on every row, which is a useful independent cross-check
that both tools' importers agree on how many distinct constraints the schema actually
defines per profile — the remaining `Generated`/`Skipped` split difference reflects a genuine
gap in codegen capability between the two tools (e.g. Dynamics: cimgo generates 1600 of the
9815, cimoxide generates 4299), not a counting discrepancy.

`-rule-report` also prints a per-file breakdown ("=== Per-File Rule Counts ===", one
`PERFILE\t<name>\t<checks>\t<skipped>\t<total>` line per TTL file) in the same format
cimoxide's `--rule-report` uses, so a profile-group mismatch between the two tools can be
localized to a specific file with no external script: `grep PERFILE cimgo.log | sort > a;
grep PERFILE cimoxide.log | sort > b; awk -F'\t' '{print $2, $5}' a | diff - <(awk -F'\t'
'{print $2, $5}' b)`.

| Profile Group | Generated | Skipped | Total |
|---------------|----------:|--------:|------:|
| Equipment (EQ) | 404 | 797 | 1201 |
| Steady State Hypothesis (SSH) | 49 | 215 | 264 |
| Dynamics (DY) | 1600 | 8215 | 9815 |
| State Variables (SV) | 52 | 90 | 142 |
| Short Circuit (SC) | 21 | 312 | 333 |
| Common / AllProfiles | 35 | 149 | 184 |
| Topology (TP) | 19 | 29 | 48 |
| DiagramLayout (DL) | 21 | 67 | 88 |
| Operation (OP) | 71 | 124 | 195 |
| **Total** | **2272** | **9998** | **12270** |

### SPARQL Check Coverage

Complex constraints defined using `sh:sparql` in the CGMES SHACL files are not automatically generated by `cmd/shaclgen`. These are instead implemented as hand-written Go functions in the `validation/` package and wired into the profile validators.

Counts below are generated by `go run ./cmd/shaclgen -rule-report`, which statically resolves the call graph in `validation/` from each profile group's entry point(s) and matches the resulting `Violation.Name` values against the distinct `sh:name`s of `sh:sparql` constraint shapes actually defined in the CGMES SHACL TTL files — re-run it after adding, removing, or renaming checks and update this table to match. Matching is done on `sh:name` (the CGMES conformance rule name, e.g. `C:452:EQ:SynchronousMachine:aggregate`) rather than the SHACL shape ID: `sh:name` is a plain string with no namespace prefix to normalize, and it's copied verbatim into `Violation.Name` on the hand-written side.

**`Implemented`/`TTL Total` count distinct named conformance rules (`sh:name` values), not distinct `sh:sparql` shapes.** A single shape's `sh:name` can itself be a `|`-joined compound of several rule names when one `sh:sparql` query enforces multiple documented conformance rules at once (e.g. one shape covering both `LoadResponseCharacteristic.exponentModel:exponent` and `:coefficient`) — both sides are split on `|` before matching, so that one shape contributes one entry to the totals per rule name it names, not one entry per shape. A shape can be partially covered: if the hand-written check only tags its `Violation.Name` with some of the rule names a shape's `sh:sparql` query is documented to enforce, the rest count as not implemented even though the shape itself has *a* check.

This table uses the same `ttlGroupLabel` grouping as "Generated SHACL Rules by Profile" above,
so their rows line up 1:1. There's no separate `C:600 conformance` row: `ValidateProf10HeaderRules`
is reached transitively from `ValidateCommonRulesSPARQL`'s call graph, so its one rule name
(`C:600:ALL:NA:PROF10`) is just counted as part of `Common / AllProfiles`.

Each manual validation rule is standardized with a header that provides traceability to the source profile and rule ID:
```go
// CheckSynchronousMachineAggregate implements eq452:SynchronousMachine-aggregate
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
```

| Profile Group | Implemented | TTL Total | Coverage |
|---------------|-------------:|----------:|---------:|
| Equipment (EQ) | 69 | 69 | 100.0% |
| Steady State Hypothesis (SSH) | 39 | 39 | 100.0% |
| Dynamics (DY) | 40 | 40 | 100.0% |
| State Variables (SV) | 11 | 11 | 100.0% |
| Short Circuit (SC) | 7 | 7 | 100.0% |
| Common / AllProfiles | 34 | 34 | 100.0% |
| Topology (TP) | 3 | 3 | 100.0% |
| DiagramLayout (DL) | 1 | 1 | 100.0% |
| Operation (OP) | 1 | 1 | 100.0% |
| **Total** | **205** | **205** | **100.0%** |

Every SPARQL constraint defined in the CGMES SHACL TTL files has a matching hand-written check.

CIMdesk quality checks (see below) have no SHACL TTL backing, so they're excluded from the TTL Total/Coverage columns above.

### CIMdesk checks outside the CGMES SHACL standard

CIMdesk implements two categories of checks that are not encoded in the CGMES SHACL TTL files and are therefore not covered by `cmd/shaclgen` or the hand-written SPARQL rules.

#### C:600 conformance rules

These carry a CGMES rule ID but are defined in the conformance document rather than the SHACL TTL files. `C:600:ALL:NA:PROF10` (file-header dependency rules) is implemented in `validation/sparql_prof10.go` and wired into `ValidateCommonRulesSPARQL`; the counts above fold it into the "Others" group. The following rule is not yet implemented:

| Rule ID | Description |
|---------|-------------|
| `C:600:ALL:NA:PROF11` | Undeclared or unrecognized classes/properties are present in the file. |

#### CIMdesk-specific modeling quality checks (no rule ID)

These have no CGMES rule ID and appear to be CIMdesk's own heuristics (14 checks total in `validation/cim_quality.go`: 13 in `ValidateCIMdeskQualityChecks` plus `CheckBaseVoltageInEQBD`, the latter carrying the informal rule ID `eqbd2:EQBD2`). They are outside the scope of the CGMES SHACL standard:

| Class | Check |
|-------|-------|
| *(global)* | No `TapChangerControl`s found — none of the `PowerTransformer`s are used for voltage regulation. |
| *(global)* | No `RegulatingControl`s found — none of the `RegulatingCondEq`s (`SynchronousMachine`, `ShuntCompensator`, `StaticVarCompensator`) are used for voltage regulation. |
| *(global)* | No boundary connections found — the IGM is an island without any inter-connections. |
| *(global)* | No `ShuntCompensator` objects found; at least one is expected. |
| `Substation` / `ControlArea` | Instance has no child objects and is not referenced by any other instance. |
| `GeographicalRegion` | None of the `PowerTransformer`s in the region are used for voltage regulation. |
| `ACLineSegment` / `DCLineSegment` | No `Location` associated with the segment. |
| `ACLineSegment` | `ACLineSegment.x / ACLineSegment.r` ratio is too large. |
| `ACLineSegment` / `PowerTransformer` | At least one associated `OperationalLimit` is violated. |
| `BaseVoltage` | Two `BaseVoltage` instances share the same `nominalVoltage` value. |
| `PowerTransformer` | Both ends of the transformer have the same `nominalVoltage`. |
| `ConnectivityNode` | Open-ended node with only one `Terminal` connected. |
| `Disconnector` | The two `ConnectivityNode`s the `Disconnector` connects are in different `VoltageLevel`s. |
| `ConformLoad` | The load and its connected `TopologicalNode`s are not in the same `EquipmentContainer`. |
| `RegulatingControl` | Target voltage deviates 10–20 % from the `nominalVoltage` of the regulated `ConnectivityNode` / `TopologicalNode`. |
| `md:FullModel` | Profile type inferred from the file contents differs from the type declared in the model header. |

