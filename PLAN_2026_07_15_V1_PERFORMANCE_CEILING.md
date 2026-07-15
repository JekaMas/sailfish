# Sailfish v1 Performance Ceiling Round

## Goal

Establish a measured same-binary implementation ceiling for every public hot
operation, close justified gaps without changing semantics or ownership, and
release the single selected implementation as `v1.0.0`.

“Ceiling” means the narrowest benchmark kernel that performs equivalent
numeric or ownership work on the same build. It is not inferred from nominal
CPU frequency.

## Matrix

| Operation family | Public workload | Measured ceiling | Candidate class | Acceptance gate |
|---|---|---|---|---|
| Runtime decimal parse | `Uint256Codec.Parse` / `ParseInto` | canonical digit kernel and complete `parseUint256` | remove redundant public-path work only | >=5% representative gain; no invalid/general regression >2% |
| Raw one-limb format | `Uint256Codec.AppendTo` | complete native formatter and fixed digit writer | remove dispatch/length work only if equivalent | >=5% gain; all scales/digit boundaries exact |
| Retained output | `Codec.AppendTo` for native and four-limb values | direct immutable-string copy | inline/dispatch only | keep only if >=5% on both lengths |
| Same-format compare | `Decimal.Compare` | direct integer compare | generic/provider dispatch only | >=5% with identical ordering |
| Checked addition | `Decimal.AddAssign` | direct limb addition | generic/type-switch dispatch only | >=5%; overflow behavior unchanged |
| CBOR scalar append/decode | runtime codec methods | direct preferred-integer codec | scale/provider dispatch only | >=5%; golden bytes and strict rejection unchanged |
| Owned output | `String`, marshal APIs | required owned result construction | no ownership weakening | no extra allocations; ownership contract remains explicit |

## Rules

- Constructors remain the only normalization/canonicalization boundary.
- Keep one production implementation for each selected operation shape.
- No compatibility path, legacy codec, fallback, alternate wire format, panic,
  floating point, unsafe, assembly, or architecture dispatch.
- Add no optimization unless repeated same-binary benchmarks, profiles,
  escape analysis, and correctness tests support it.
- Revert rejected candidates completely and record the result in
  `PERFORMANCE.md`.

## Gates

```sh
GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench '^BenchmarkPerformanceCeilings$' -benchmem -benchtime=500ms -count=15
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off go test -race ./...
GOWORK=off make fuzz
GOWORK=off make bench
```

The release commit must be clean under `git diff --check`, contain no rejected
implementation, and be tagged once as `v1.0.0` after push verification.

## Result

| Action | Decision | Evidence |
|---|---|---|
| Direct one-byte runtime scale | Keep | Parse 9.80 -> 8.75 ns; complete parser 8.63 ns |
| Scale-independent runtime CBOR wrappers | Keep | Decode 6.74 -> 4.24 ns; complete decoder 4.23 ns |
| Value-only parse helper | Reject and remove | Direct helper improved; public method exceeded inline budget and did not |
| Generic type-switch compare | Reject and remove | Wide compare regressed from about 5.5 to 8.4 ns |
| Provider-dispatched checked add | Reject and remove | No stable gain across both native and wide backends |
| Unsafe, assembly, pooling, alternate wire | Not applicable | Hot paths are 0 alloc and selected operations reached their pure-Go kernels |

`make test`, `make vet`, `make race`, all fuzz targets, and the full five-run
`make bench` matrix passed. Detailed measurements and compiler/profile owners
are recorded in `PERFORMANCE.md` and `BENCHMARKS.md`.
