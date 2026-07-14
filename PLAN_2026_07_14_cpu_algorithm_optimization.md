# Sailfish CPU And Algorithm Optimization Plan

**Status:** IN PROGRESS

## Goal

Run one evidence-backed scalar optimization round over Sailfish decimal parsing
and formatting. Try candidates independently, retain only measured wins, and
preserve exact fixed-scale semantics, zero-allocation append/parse hot paths,
and the public API.

## Protected Behavior

- Parsing accepts and rejects exactly the same syntax, precision, and range
  cases as the current implementation.
- Formatting remains canonical and byte-for-byte identical.
- Constructors remain the only normalization/canonicalization boundary.
- `ParseCompact`, `ParseBytes`, and pre-sized `AppendTo` remain allocation-free.
- No unsafe, assembly, architecture dispatch, mutable string aliasing, or
  public API/storage-format changes in this round.

## Benchmark Gate

Use Go 1.26.5 on darwin/arm64 with `GOWORK=off`, fixed benchmark regexes,
`-benchmem`, at least 10 repeated runs, CPU profiles for time changes, and
escape analysis for allocation changes. Compare with `benchstat` or equivalent
repeated-run summaries.

Keep a candidate only when all of these hold:

1. It improves a representative target by at least 5% or removes measurable
   work identified by CPU profiling.
2. It does not regress another protected representative path by more than 2%
   without a documented workload justification.
3. Allocations do not increase; hot append/parse paths stay at 0 B/op and
   0 allocs/op.
4. Unit, property, fuzz-seed, allocation, and full package tests pass.
5. The implementation is simpler than, or comparably maintainable to, the
   measured benefit.

Rejected candidates are removed. Successful candidates are recorded as a
decision in `PERFORMANCE.md`, committed independently, and pushed.

The baseline and every candidate comparison use `GOMAXPROCS=1` to reduce
scheduler noise. A candidate is compared with its current implementation in
the same benchmark binary and run. Wall-clock results are summarized from ten
runs; allocation counts must agree in every run. CPU profiles, escape analysis,
and bounds-check diagnostics identify the changed owner before a production
edit is accepted.

## Workload Matrix

| Dimension | Required cases |
|---|---|
| Native parse length | 1, 2, 7, 8, 9, 16, 19, and 20 digits |
| Native fixed-scale parse | scales 0, 2, 5, 9, and 18; integer and fraction on both sides of the point |
| Wide parse length | 19, 20, 38, 39, 57, 58, 77, and overflow-length digits |
| Input ownership | `string`, `[]byte`, retained canonical, compact canonical |
| Invalid input | first/middle/last invalid byte, duplicate point, excess precision, and overflow |
| Batch parse | realistic mixed prices/amounts in batches of 1, 8, 64, and 256 values |
| Native format | every decimal digit width 1-20; scales below/equal/above digit count |
| Wide format | one-, two-, three-, and four-limb values; scales 0, 5, and 18 |
| Output ownership | retained append, raw-unit append, pre-sized destination, undersized destination |

## Evidence Matrix

| Evidence | Baseline | Every retained candidate | Final round |
|---|---:|---:|---:|
| Repeated `-benchmem` results | Required | Required before/after | Required |
| CPU profile | Required | Required for changed owner | Required |
| Allocation profile | Required | Required if allocation shape changes | Required |
| Escape analysis | Required | Required | Required |
| Bounds-check diagnostics | Required | Required for parser/formatter kernels | Required |
| Unit/property/allocation tests | Required | Required | Required |
| Race/vet/fuzz | Required | Required before commit | Required |

## Candidate Matrix

| ID | Candidate | Target workloads | Required comparison | Decision |
|---|---|---|---|---|
| C1 | Eight-digit SWAR validation/parsing with scalar tail | canonical `uint64` and `uint256` strings/bytes at 8, 16, 19, 38, and 77 digits | current pair parser vs SWAR; valid, syntax error, overflow, string, bytes, and batch inputs | Pending |
| C2 | `uint64` formatter variants: current divide-by-100, `1e9` outer chunks, and direct fixed-scale placement | 1-20 digits; scales 0, 2, 5, 9, 18; units below/equal/above scale | length-distributed append benchmarks and CPU profiles | Pending |
| C3 | `uint256` decimal chunk base `1e9` vs current `1e19` | 65-, 128-, 192-, and 256-bit values; scales 0, 5, 18 | split-only and full append benchmarks; division count and CPU profiles | Pending |
| C4 | Remove decimal-point prefix copy by formatting integer/fraction regions directly | common prices and wide amounts where digits exceed scale | current copy path vs direct placement across scale/digit relationships | Pending |
| C5 | Small-value formatting cache | repeated 0-99/0-999 units and representative market values | hit/miss distribution benchmarks and retained-size review | Pending only if profiles show relevant formatting cost |
| C6 | Optional branch-minimized/SWAR validation refinements | batch parsing of realistic market-data decimal distributions | scalar vs candidate batch throughput and code-size review | Pending only after C1-C4 |

## Candidate Actions

| Candidate | Prototype action | Keep condition | Reject condition |
|---|---|---|---|
| C1 | Add benchmark-only eight-digit SWAR string/byte kernels, then integrate into the existing parser only if they win | At least 5% on dense 8-77 digit and batch workloads with no short/error-path regression above 2% | Smaller win, code-size/BCE regression, or any semantic/allocation change |
| C2 | Compare current divide-by-100 writer with `1e9` outer chunks and direct fixed-width placement | Representative native formatting improves at least 5% and every digit/scale property remains exact | Narrow synthetic-only win or regression on common 5/9-scale prices |
| C3 | Compare current `1e19` wide chunks with `1e9`, measuring split and full append independently | Wide representative append improves at least 5% without parse/size regression | More divisions or worse one-/two-limb behavior |
| C4 | Write digits directly around the decimal point without prefix copy | Values with digits greater than scale improve at least 5%, no regression elsewhere | Compiler already optimizes the copy or direct placement adds branches/work |
| C5 | Measure a small-value table without shipping it first | Profile demonstrates relevant repeated small raw-unit formatting and aggregate benefit exceeds retained data/code cost | Profile does not justify it or cache only helps artificial distributions |
| C6 | Measure branch-minimized validation after C1-C4 | Mixed valid/error batches improve at least 5% with equivalent diagnostics | Valid hot path regresses or code obscures BCE/inlining |

## Test Matrix

| Surface | Cases |
|---|---|
| Parse correctness | zero, leading zeros, integer-only, exact fraction, short fraction, excess precision, invalid byte, multiple dots, empty input, max value, overflow |
| Parse cross-products | `string` and `[]byte`; uint8/16/32/64/256; scales 0-20 and representative wide scales |
| Format correctness | every digit length for native widths; zero; powers of ten boundaries; scale below/equal/above digit count; maximum values |
| Round trip | format(parse(x)) and parse(format(units)) properties for every backend |
| Allocation | compact string parse, byte parse, pre-sized append, CBOR append/parse |
| Regression | `go test ./...`, race test, vet, formatting, and clean git diff |

## Execution Order

1. Capture baseline benchmarks, CPU profiles, allocation profiles, and escape
   analysis for current parse/format kernels.
2. Add distribution benchmarks that are missing from the existing suite.
3. Run C1 through C4 one at a time. Revert rejected prototypes completely.
4. Run C5/C6 only if profiles and prior results justify them.
5. For every accepted candidate: update `PERFORMANCE.md`, run all gates, make a
   self-contained commit, and push.
6. Finish with a full benchmark comparison and package-wide verification.

## Commit Policy

- Commit and push each retained production optimization with its
  `PERFORMANCE.md` decision and before/after evidence.
- Benchmark-only harness improvements may be grouped in one initial commit.
- Do not commit rejected implementation variants, alternate runtime paths,
  compatibility code, or architecture fallbacks.
- `main` always contains one selected implementation per operation.
