# CIMgo

CIMgo processes CGMES/CIM XML/RDF files in Go and protobuf.

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

Across 73 profiles there are 9153 constraints total. Of these, 6971 generate
code and 2182 are skipped: 2146 are structurally satisfied by the Go type
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

### Structurally satisfied (2146 skips)

| Count | Constraint | Reason |
|------:|-----------|--------|
| 1302 | `sh:maxCount 1` on scalar fields | Scalar fields (`int`, `float`, `bool`, `string`) hold exactly one value; MaxCount ≥ 1 is vacuously true. |
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
