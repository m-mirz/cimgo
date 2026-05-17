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

Merges one or more CGMES XML files and outputs the combined dataset as JSON.

```bash
cimcli convert <xml-file1> [<xml-file2> ...]
```

### Examples

**Validate a full model:**
```bash
# Linux
./cimcli-linux-amd64 validate EQ.xml SSH.xml TP.xml SV.xml

# Windows
cimcli-windows-amd64.exe validate EQ.xml SSH.xml TP.xml SV.xml
```

**Validate specific profiles:**
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
./cimcli-linux-amd64 validate -json \
  20210401T1730Z_1D_BE_EQ_1.xml \
  20210401T1730Z_1D_BE_SSH_1.xml \
  20210401T1730Z_1D_BE_TP_1.xml \
  20210401T1730Z_1D_BE_SV_1.xml
```

**Convert files to JSON:**
```bash
./cimcli-linux-amd64 convert EQ.xml SSH.xml > dataset.json
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

In the flame graph, the widest bands are the hottest call stacks. The **Top**
view's `flat` column shows a function's own time; `cum` includes its callees.
Click any function in **Top** to open the annotated **Source** view.

## Architecture

The code generation process follows these main steps:

1.  **Schema Loading:** The tool begins by finding and parsing the relevant CIM RDF schema or TTL SHACL files based on the specified version and profiles.
2.  **Schema Processing:** It processes the parsed RDF or SHACL data into an internal, language-agnostic representation of CIM classes, properties, datatypes, their relationships and rules.
3.  **Code Generation:** Using Go's `text/template` engine, it feeds the internal representation into language-specific templates (`lang-templates/*.tmpl`) to generate the final source code files.

## SHACL Validation

`cmd/shaclgen` translates each constraint in the CGMES SHACL Turtle files
into a Go `Check<...>` function under `shaclgen/`. The generated checks are
wired into per-profile orchestrators and aggregated by
`shaclgen.ValidateAllGeneratedProfiles`.

For a complete validation pass including both generated and hand-written SPARQL
rules, use `validation.ValidateAllProfiles`.

Across 73 profiles there are 9153 constraints total. Of these, 4184 generate
code and 4969 are skipped: 4933 are structurally satisfied by the Go type
system (generating a check would never fire or would produce false positives)
and 36 cannot be conducted due to upstream SHACL TTL defects.

### Simplification rules applied during import

Before code generation, `shaclimport.SimplifyFileResults` normalises each
property shape's constraint list. Rules are applied in order; a constraint
that matches a rule is either dropped or rewritten and is not passed on to
`cmd/shaclgen`.

| Rule | Constraint removed / rewritten | Reason |
|------|-------------------------------|--------|
| 1 | `sh:nodeKind sh:Literal` when any `sh:datatype` is also present | `sh:datatype` already implies the value is a literal; the `sh:nodeKind` check is redundant. |
| 2 | `sh:nodeKind sh:BlankNodeOrIRI` and `sh:nodeKind sh:IRI` unconditionally | Every IRI-typed CIM property is generated as a Go reference field (`*struct{ MRID string }`); the type system already enforces the IRI shape. |
| 3 | `sh:minCount 0` | Vacuously true — zero or more values are always acceptable. |
| 4 | `sh:datatype xsd:T` for native Go scalar types | The Go struct field is already typed (`int`, `float64`, `bool`, `string`), so the XML decoder rejects malformed input before validation. Dropped for: all integer variants, `float`/`double`/`decimal`, `boolean`, `string`/`normalizedString`/`token`. Non-native types (`dateTime`, `gMonthDay`, `anyURI`) are **not** dropped. |
| 5 | `sh:in` with a single value → rewritten as `sh:hasValue` | A one-element allow-list is semantically identical to an exact-value check. |
| 6 | `sh:minCount 0` + `sh:maxCount 1` → synthetic `sh:Optional` | The pair means "0 or 1 values". `sh:minCount 0` is dropped by Rule 3; the matching `sh:maxCount 1` is replaced by a single `Optional` sentinel to record the upper bound without implying a presence requirement. |
| 7 | `sh:minCount 1` + `sh:maxCount 1` → synthetic `sh:Required` | The pair means "exactly 1 value". Both constraints are collapsed into a single `Required` sentinel, avoiding duplicate presence checks. |

### Structurally satisfied (4948 skips)

| Count | Constraint | Reason |
|------:|-----------|--------|
| 1302 | `sh:maxCount 1` on scalar fields | Scalar fields (`int`, `float`, `bool`, `string`) hold exactly one value; MaxCount ≥ 1 is vacuously true. |
| 2787 | `sh:required` on `float` fields | Float fields use `omitempty`; `0.0` is indistinguishable from absent after XML decode. This makes presence checks unreliable: a legitimately-zero physical quantity (e.g. `bch=0`, `r=0`, `b=0`) would always trigger a false positive. Range constraints (`sh:minExclusive`, `sh:minInclusive`) cover the must-be-positive subset where zero is genuinely invalid. Fixing the general case would require switching all float fields to `*float64`. |
| 413 | `sh:required` on `bool` fields | Bool fields use `omitempty`; `false` is indistinguishable from absent after XML decode. Fixing would require switching all bool fields to `*bool`. |
| 334 | `sh:maxCount 1` on pointer fields | Pointer fields are either nil (0 values) or non-nil (1 value). |
| 56 | `sh:maxCount 1` on multi-hop paths | Every hop in a CIM reference path is a 0..1 pointer, so the count is always ≤ 1. |
| 18 | Cross-class `sh:lessThan` on sibling subtypes | The SHACL property shape reuses the same comparison across multiple target classes. The compared field (`xDirectSubtrans`, `xQuadSubtrans`, `xpp`) exists only on a sibling subtype (`SynchronousMachineTimeConstantReactance` or `AsynchronousMachineTimeConstantReactance`). These subtypes are mutually exclusive in valid CGMES data — a machine is either `SynchronousMachineTimeConstantReactance` or `SynchronousMachineEquivalentCircuit`, never both — so both operands can never be visible at the same time. The same-class cases (`SynchronousMachineTimeConstantReactance.statorLeakageReactance < xDirectSubtrans`, etc.) are generated normally. |
| 15 | Inverse `sh:class` | The asserted class is an ancestor of every concrete target subclass; Go struct embedding guarantees the constraint is always satisfied. |
| 11 | `sh:nodeKind` on `rdf:type` paths | The Go struct type is fixed at decode time, so the RDF type is always correct. |
| 2 | `sh:datatype xsd:anyURI` on slice fields | A slice field (e.g. `Profile []string`) holds zero or more values; per-element URI format validation requires iterating the slice, which `shaclgen` does not generate for datatype checks. URI format is not validated by the Go XML decoder either, so this is an acknowledged gap rather than a false negative. |
| 4 | `sh:in (dm:DifferenceModel md:FullModel)` on `Supersedes`/`DependentOn` | The constraint checks that each referenced model's `rdf:type` is either `FullModel` or `DifferenceModel`. `CIMElementList` stores models in two typed maps (`FullModels map[string]*FullModel`, `DifferenceModels map[string]*DifferenceModel`); any MRID that resolves within the dataset is guaranteed to be in one of those two types. |
| 3 | `sh:hasValue rdf:type rdf:Statements` on `forwardDifferences`/`reverseDifferences`/`preconditions` | The CGMES difference model format mandates that all elements in these collections are `rdf:Statement` resources. The referenced objects are not decoded into `CIMElementList` (they are RDF graph metadata, not CIM elements), so the constraint cannot be checked at runtime — but it cannot be violated by any well-formed CGMES difference model file. |
| 3 | Multi-segment `sh:required` on `rdf:Statements.subject/predicate/object` | The RDF specification mandates that every `rdf:Statement` resource has `subject`, `predicate`, and `object` predicates. Any `rdf:Statement` instance loaded from a valid RDF document already satisfies these constraints by definition. |

### Cannot be conducted (22 skips)

#### Upstream SHACL TTL defects

**Field name typos** — `sh:lessThan` references a misspelled field name; all four are in `61970-302_Dynamics-AP-Con-Complex-SHACL.ttl`:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `C:302:DY:GovHydroIEEE0.pmin:valueRangePair` | `sh:lessThan cim:GovHydroIEEE.pmax` — class suffix `0` missing; should be `GovHydroIEEE0.pmax` |
| `C:302:DY:PFVArType1IEEEVArController.vvtmin:valueRangePair` | `sh:lessThan cim:PVFArType1IEEEVArController.vvtmax` — prefix transposed; should be `PFVArType1IEEEVArController.vvtmax` |
| `C:302:DY:ExcDC1A.efdmin:valueRangePair` | `sh:lessThan cim:ExcDC1A.edfmax` — letters transposed; should be `efdmax` |
| `C:302:DY:PssIEEE4B.vhmin:valueRangePair` | `sh:lessThan cim:PssIEEE4V.vhmax` — class suffix wrong; should be `PssIEEE4B.vhmax` |

**Class name typo** — one property shape in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` references a misspelled class in its inverse path, generating 4 skip instances (one per concrete target class):

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `SynchronousMachineDynamics.CrossCompoundTurbineGovernorDyanmics-cardinality` | `sh:inversePath cim:CrossCompoundTurbineGovernorDyanmics.SynchronousMachineDynamics` — "Dynamics" misspelled as "Dyanmics"; applied to `SynchronousMachineEquivalentCircuit`, `SynchronousMachineSimplified`, `SynchronousMachineTimeConstantReactance`, `SynchronousMachineUserDefined` |

**Class name capitalisation mismatch** — two SHACL files (`61970-457_Dynamics-AP-Con-Complex-Explicit-CrossProfile-SHACL.ttl` and `61970-457_Dynamics-AP-Con-Complex-Implicit-CrossProfile-SHACL.ttl`) reference `cim:CSConverter` (capital S), but all RDFS schema files consistently define the class as `cim:CsConverter`. The two defective shapes generate 2 skip instances:

| Shape | Defect |
|-------|--------|
| `dy457cpe:CSCDynamics.CSConverter-valueType` | `sh:in (cim:CSConverter)` — should be `cim:CsConverter` |
| `dy457cpi:CSCDynamics.CSConverter-valueType` | same defect in the implicit cross-profile file |

**Wrong field names in inverse paths** — four property shapes in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` reference field names that do not match the RDFS schema, generating 8 skip instances:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `SynchronousMachineDynamics.CrossCompoundTurbineGovernorDynamics-cardinality` | `sh:inversePath cim:CrossCompoundTurbineGovernorDynamics.SynchronousMachineDynamics` — no such field; RDFS defines `HighPressureSynchronousMachineDynamics` and `LowPressureSynchronousMachineDynamics`; applied to 4 concrete target classes |
| `CsConverter.CSCDynamics-cardinality` | `sh:inversePath cim:CSCDynamics.CsConverter` — capitalisation wrong; RDFS defines `CSCDynamics.CSConverter` |
| `VCompIEEEType2.GenICompensationForGenJ-cardinality` | `sh:inversePath cim:GenICompensationForGenJ.VCompIEEEType2` — capitalisation wrong; RDFS defines `GenICompensationForGenJ.VcompIEEEType2` |
| `WindContQIEC.WindTurbineType3or4IEC-cardinality` | `sh:inversePath cim:WindTurbineType3or4IEC.WindContQIEC` — capitalisation wrong; RDFS defines `WindTurbineType3or4IEC.WIndContQIEC` |

**Stale field reference** — one property shape in `61970-301_Operation-AP-Con-Complex-SHACL.ttl` references a field that was present in CGMES 2.4 but removed from the current CGMES 3.0 schema:

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `C:301:OP:AccumulatorValue.value:valueRange` | `sh:minExclusive` on `cim:AccumulatorValue.value` — field removed in CGMES 3.0 |

**Non-existent target class** — one property shape in `61970-600-2_Dynamics-AP-Con-Complex-InverseAssociation-SHACL.ttl` lists `cim:GovHydroIEEE1` in its `sh:targetClass` alongside several real classes. No such class exists in the CIM standard or in `cimstructs`; `shaclgen` silently skips it when resolving concrete target classes.

**Empty `sh:in` list** — one property shape in `61970-600-2_Operation-AP-Con-Simple-SHACL.ttl` has `sh:in ()`, generating 4 skip instances (one per concrete target class):

| Rule (`sh:name`) | Defect |
|------------------|--------|
| `Measurement.Terminal-valueType` | `sh:in ()` — empty allow-list; applied to `Accumulator`, `Analog`, `Discrete`, `StringMeasurement` |

#### Generator limitations

Nine shapes in the SHACL files use `sh:targetNode <sentinel> + sh:sparql` to express constraints that span many instances or require graph traversal: `AllGeneratingUnit`, `AngleReference`, `DanglingReferences`, `FloatSpecialValues`, `IDchecks`, `IDuniqueness`, `IdentifiedObjectStringLength`, `SubstationCount`, `TextDiagramObjectDiagramObject`. None of these names correspond to a real CIM class, so `shaclgen` silently produces no output for them. All nine are instead implemented as hand-written Go functions in `validation/` (see SPARQL Check Coverage below).

### SPARQL Check Coverage

Complex constraints defined using `sh:sparql` in the CGMES SHACL files are not automatically generated by `cmd/shaclgen`. These are instead implemented as hand-written Go functions in the `validation/` package and wired into the profile validators.

Across all profiles, there are approximately **200** SPARQL-based constraints. All active constraints are currently implemented, covering 100% of the active validation requirements defined in the CGMES standard.

Each manual validation rule is standardized with a header that provides traceability to the source profile and rule ID:
```go
// CheckSynchronousMachineAggregate implements eq452:SynchronousMachine-aggregate
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
```

| Profile Group | SPARQL Constraints | Implemented | Coverage |
|---------------|-------------------:|------------:|---------:|
| Equipment (EQ) | 66 | 66 | 100% |
| Steady State Hypothesis (SSH) | 40 | 40 | 100% |
| Dynamics (DY) | 40 | 40 | 100% |
| State Variables (SV) | 11 | 11 | 100% |
| Short Circuit (SC) | 7 | 7 | 100% |
| Others (TP, DL, All, etc.) | 28 | 28 | 100% |
| **Total** | **192** | **192** | **100%** |

### CIMdesk checks outside the CGMES SHACL standard

CIMdesk implements two categories of checks that are not encoded in the CGMES SHACL TTL files and are therefore not covered by `cmd/shaclgen` or the hand-written SPARQL rules.

#### C:600 conformance rules not yet implemented

These carry a CGMES rule ID but are defined in the conformance document rather than the SHACL TTL files:

| Rule ID | Description |
|---------|-------------|
| `C:600:ALL:NA:PROF11` | Undeclared or unrecognized classes/properties are present in the file. |
| `C:600:EQ:Substation:count` | The number of `Substation`s shall be less than the number of `VoltageLevel`s; each `Substation` should contain more than one `VoltageLevel`. |

#### CIMdesk-specific modeling quality checks (no rule ID)

These have no CGMES rule ID and appear to be CIMdesk's own heuristics. They are outside the scope of the CGMES SHACL standard:

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

