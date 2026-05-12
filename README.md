# CIMgo

CIMgo processes CGMES/CIM XML/RDF files in Go and protobuf.

## cimval — CGMES Validation CLI

`cimval` validates CGMES XML instance files against the CGMES SHACL and SPARQL rules.

### Download

Pre-built binaries are attached to each [GitHub release](https://github.com/m-mirz/cimgo/releases/latest):

| Platform | File |
|----------|------|
| Linux (x86-64) | `cimval-linux-amd64` |
| Windows (x86-64) | `cimval-windows-amd64.exe` |

**Linux** — make the binary executable after downloading:

```bash
chmod +x cimval-linux-amd64
```

**Windows** — the `.exe` can be run directly from PowerShell or CMD.

### Usage

Pass one or more CGMES XML files. Profiles, solved/not-solved state, and
EQBD base voltage IDs are detected automatically from the file headers:

```bash
# Linux
./cimval-linux-amd64 EQ.xml SSH.xml TP.xml SV.xml

# Windows
cimval-windows-amd64.exe EQ.xml SSH.xml TP.xml SV.xml
```

**Options:**

```
-profile  Comma-separated list of profiles to check (EQ,SSH,TP,SV,DY,SC,DL,GL,OP,EQBD).
          Default: auto-detected from file headers.
-solved   Enable SolvedMAS checks (default: auto-detected from SV profile presence).
-notsolved Enable NotSolvedMAS checks (default: auto-detected).
-common   Enable Common/AllProfiles rules (default: true).
-quality  Enable CIMdesk-style modeling quality checks (default: false).
-silence  Comma-separated list of rule IDs to suppress.
-json     Output violations as JSON instead of plain text.
```

Exit code is `0` when no `sh:Violation`-severity findings are present, `1` otherwise.

### Example

Validate a MicroGrid MAS dataset and show violations as JSON:

```bash
./cimval-linux-amd64 -json \
  20210401T1730Z_1D_BE_EQ_1.xml \
  20210401T1730Z_1D_BE_SSH_1.xml \
  20210401T1730Z_1D_BE_TP_1.xml \
  20210401T1730Z_1D_BE_SV_1.xml
```

---

## How to Build and Run

Make sure that you have cloned the repo recursively to include the CGMES schema files from ENTSO-E

    git clone --recurse-submodules [...]

or clone the submodule in a second step

    git submodule update --init --recursive

Ensure that GOPATH is set and included in your PATH.

For the protobuf code generation, you also require the proto compiler

    sudo apt-get install -y protobuf-compiler
    # install tools from mod file
    go get tool

First, you need to generate the cim based code to be able to build the entire package.

```bash
go generate ./...
```

## How to Test

Run the test suite using the `go test` command. The `-v` flag provides verbose output.

    go test -v ./...


## Architecture

The code generation process follows these main steps:

1.  **Schema Loading:** The tool begins by finding and parsing the relevant CIM RDF schema files based on the specified version and profiles.
2.  **Schema Processing:** It processes the parsed RDF data into an internal, language-agnostic representation of CIM classes, properties, datatypes, and their relationships.
3.  **Code Generation:** Using Go's `text/template` engine, it feeds the internal representation into language-specific templates (`lang-templates/*.tmpl`) to generate the final source code files.

## Key Go Files

*   `cmd/cimgen/main.go`: The main entry point for the CLI tool. It parses command-line arguments and orchestrates the code generation process.
*   `cim_generate.go`: Contains the core logic for driving the generation process for a specific language.
*   `cim_schema_import.go`: Handles the discovery and parsing of the CIM RDF schema files.
*   `cim_schema_processing.go`: Responsible for transforming the raw parsed schema into the internal representation used by the generators.
*   `generator_*.go`: A set of files (e.g., `generator_go.go`, `generator_cpp.go`) that implement the generation logic for each target language.
*   `templates.go`: Manages the embedded template files.

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

### Structurally satisfied (4933 skips)

| Count | Constraint | Reason |
|------:|-----------|--------|
| 1302 | `sh:maxCount 1` on scalar fields | Scalar fields (`int`, `float`, `bool`, `string`) hold exactly one value; MaxCount ≥ 1 is vacuously true. |
| 2787 | `sh:required` on `float` fields | Float fields use `omitempty`; `0.0` is indistinguishable from absent after XML decode. This makes presence checks unreliable: a legitimately-zero physical quantity (e.g. `bch=0`, `r=0`, `b=0`) would always trigger a false positive. Range constraints (`sh:minExclusive`, `sh:minInclusive`) cover the must-be-positive subset where zero is genuinely invalid. Fixing the general case would require switching all float fields to `*float64`. |
| 413 | `sh:required` on `bool` fields | Bool fields use `omitempty`; `false` is indistinguishable from absent after XML decode. Fixing would require switching all bool fields to `*bool`. |
| 334 | `sh:maxCount 1` on pointer fields | Pointer fields are either nil (0 values) or non-nil (1 value). |
| 56 | `sh:maxCount 1` on multi-hop paths | Every hop in a CIM reference path is a 0..1 pointer, so the count is always ≤ 1. |
| 18 | Cross-class `sh:lessThan` on sibling subtypes | The SHACL property shape reuses the same comparison across multiple target classes. The compared field (`xDirectSubtrans`, `xQuadSubtrans`, `xpp`) exists only on a sibling subtype (`SynchronousMachineTimeConstantReactance` or `AsynchronousMachineTimeConstantReactance`). These subtypes are mutually exclusive in valid CGMES data — a machine is either `SynchronousMachineTimeConstantReactance` or `SynchronousMachineEquivalentCircuit`, never both — so both operands can never be visible at the same time. The same-class cases (`SynchronousMachineTimeConstantReactance.statorLeakageReactance < xDirectSubtrans`, etc.) are generated normally. |
| 15 | Inverse `sh:class` | The asserted class is an ancestor of every concrete target subclass; Go struct embedding guarantees the constraint is always satisfied. |
| 8 | `sh:nodeKind` on `rdf:type` paths | The Go struct type is fixed at decode time, so the RDF type is always correct. |

### Cannot be conducted (36 skips)

#### Upstream SHACL TTL defects

| Count | Kind | Detail |
|------:|------|--------|
| 4 | Field name typo | `ExcDC1A.edfmax` (→ `efdmax`), `GovHydroIEEE.pmax` (→ `GovHydroIEEE0.pmax`), `PVFArType1IEEEVArController.vvtmax` (→ `PFVArType1...`), `PssIEEE4V.vhmax` (→ `PssIEEE4B.vhmax`) |
| 4 | Class name typo | `CrossCompoundTurbineGovernorDyanmics` (extra 'a' in "Dynamics") used as inverse target class |
| 4 | Empty `sh:in` list | `sh:in ()` with no values in inverse-association profiles |
| 4 | Non-`cim:` namespace | `mdc:FullModel`, `diff:DifferenceModel` — Header and AllProfiles shapes outside the CIM namespace |

#### Fields or classes absent from `cimgostructs`

| Count | Kind | Detail |
|------:|------|--------|
| 8 | Field missing from struct | `CrossCompoundTurbineGovernorDynamics.SynchronousMachineDynamics` (4), `AccumulatorValue.value`, `CSCDynamics.CsConverter`, `GenICompensationForGenJ.VCompIEEEType2`, `WindTurbineType3or4IEC.WindContQIEC` |
| 2 | Class name capitalisation mismatch | SHACL uses `CSConverter`; `cimgostructs` generates `CsConverter` |
| 10 | Class not generated | `AllGeneratingUnit`, `AngleReference`, `DanglingReferences`, `FloatSpecialValues`, `GovHydroIEEE1`, `IDchecks`, `IDuniqueness`, `IdentifiedObjectStringLength`, `SubstationCount`, `TextDiagramObjectDiagramObject` |

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

