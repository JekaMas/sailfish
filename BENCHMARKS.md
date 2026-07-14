# Performance Notes

## Environment

```text
Go:    go1.26.5
OS:    darwin
Arch:  arm64
CPU:   Apple M1 Max
Module: github.com/JekaMas/sailfish
```

Benchmark numbers are local measurements, not portable guarantees. Compare
changes using the same host, Go version, benchmark regex, workload, and run
count.

## Commands

```sh
GOWORK=off go test -run '^$' \
  -bench 'Benchmark(CodecParse|AppendTo|String|Compare|AddAssign)' \
  -benchmem -benchtime=200ms -count=10

GOWORK=off go test -run '^$' \
  -bench 'BenchmarkCodecParse/uint256/canonical$|BenchmarkAppendTo/uint256/formatted$' \
  -benchtime=2s -cpuprofile cpu.pprof

GOWORK=off go test -run '^$' \
  -bench 'BenchmarkString/uint64/formatted$' \
  -benchtime=2s -memprofile alloc.pprof -memprofilerate=1

GOWORK=off go test -run '^$' -gcflags='all=-m=2' 2> escape.txt

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench '^BenchmarkUint256MarketHotPaths$' \
  -benchmem -benchtime=1s -count=10

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench 'Benchmark(ScaleMetadataDispatch|GenericBackendDispatch|ExplicitUnitWidths)$' \
  -benchmem -benchtime=500ms -count=10
```

## Snapshot

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `CodecParse/uint64/canonical` | 9.79 ns/op | 0 | 0 |
| `CodecParse/uint64/compact` | 9.55 ns/op | 0 | 0 |
| `CodecParse/uint64/bytes` | 9.77 ns/op | 0 | 0 |
| `CodecParse/uint64/invalid` | 10.6 ns/op | 0 | 0 |
| `CodecParse/uint256/canonical` | 51.7 ns/op | 0 | 0 |
| `CodecParse/uint256/max` | 80.3 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/parse_into_runtime_codec/cex_scale6_one_limb` | 10.2 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/cex_scale6_one_limb` | 3.52 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/scale18_four_limb` | 4.22 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_runtime_codec/cex_scale6_one_limb` | 13.1 ns/op | 0 | 0 |
| `AppendTo/uint64/retained` | 2.73 ns/op | 0 | 0 |
| `AppendTo/uint64/formatted` | 13.7 ns/op | 0 | 0 |
| `AppendTo/uint256/formatted` | 152 ns/op | 0 | 0 |
| `String/uint64/retained` | 2.14 ns/op | 0 | 0 |
| `String/uint64/formatted` | 32.8 ns/op | 16 | 1 |
| `Compare/uint64/same-scale` | 2.14 ns/op | 0 | 0 |
| `Compare/uint256/same-scale` | 6.57 ns/op | 0 | 0 |
| `Compare/cross-scale` | 52.7 ns/op | 0 | 0 |
| `AddAssign/uint64` | 4.48 ns/op | 0 | 0 |
| `AddAssign/uint256` | 13.8 ns/op | 0 | 0 |
| `ReferenceStrconvSplitUint64` | 19.3 ns/op | 0 | 0 |

## CBOR

Sailfish encodes a decimal as its scaled unsigned integer. This is the compact
scalar representation used when the decimal is embedded in a parent
`cbor:",toarray"` record. Scale and retained source text are not serialized.

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `CBORDispatchLayers/append/codec_uint64` | 4.73 ns/op | 0 | 0 |
| `CBORDispatchLayers/decode/codec_uint64` | 5.08 ns/op | 0 | 0 |
| `CBORDispatchLayers/append/decimal_uint64` | 7.51 ns/op | 0 | 0 |
| `CBORDispatchLayers/decode/decimal_uint64` | 8.55 ns/op | 0 | 0 |
| `CBORUint256Widths/runtime_codec_append/one_limb` | 4.21 ns/op | 0 | 0 |
| `CBORUint256Widths/runtime_codec_append/maximum` | 6.91 ns/op | 0 | 0 |
| `CBORUint256Widths/runtime_codec_decode/one_limb` | 7.27 ns/op | 0 | 0 |
| `CBORUint256Widths/runtime_codec_decode/maximum` | 9.29 ns/op | 0 | 0 |
| `CBORToArrayIntegration/marshal` | 206 ns/op | 120 | 4 |
| `CBORToArrayIntegration/unmarshal` | 154 ns/op | 0 | 0 |
| `CBORManualPositionalBar/encode` | 55.5 ns/op | 0 | 0 |
| `CBORManualPositionalBar/decode` | 112.7 ns/op | 8 | 1 |

The direct caller-buffer path is the MDBX hot path. `MarshalCBOR` must return
an owned slice and therefore has one required result allocation. Reflective
`fxamacker` aggregate marshal additionally invokes that interface for each
decimal; its allocations are library/interface ownership, not decimal parsing
or unit conversion. Aggregate decode and direct strict decode are allocation-
free in the measured fixed-array cases. The fourteen-field manual positional
benchmark uses the same 93-byte wire as deterministic fxamacker `toarray`.
Its one decode allocation owns the parent record's symbol string; all four
price fields and the wide amount decode without allocation through
`ParseCBORFirst`.

CPU profiles place direct CBOR work in shortest-form integer validation,
`uint256` byte-width selection, big-endian limb transfer, and scale validation.
Memory profiles contain no per-operation allocation from append or decode.
Cached `Codec` methods avoid repeated generic scale metadata work and are the
recommended hot-loop surface. The figures above are means of ten
`GOMAXPROCS=1`, 200 ms runs on the documented host.

## Scale and backend composition

Fractional scale does not imply storage width. Explicit `uint8`, `uint16`,
`uint32`, and `uint64` formats all remain allocation-free:

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `ExplicitUnitWidths/parse/uint8` | 8.86 ns/op | 0 | 0 |
| `ExplicitUnitWidths/parse/uint16` | 8.64 ns/op | 0 | 0 |
| `ExplicitUnitWidths/parse/uint32` | 8.65 ns/op | 0 | 0 |
| `ExplicitUnitWidths/parse/uint64` | 7.66 ns/op | 0 | 0 |
| `ExplicitUnitWidths/append/uint8` | 10.7 ns/op | 0 | 0 |
| `ExplicitUnitWidths/append/uint16` | 10.7 ns/op | 0 | 0 |
| `ExplicitUnitWidths/append/uint32` | 10.6 ns/op | 0 | 0 |
| `ExplicitUnitWidths/append/uint64` | 10.8 ns/op | 0 | 0 |

Generic semantic-kind/scale composition has no heap cost. Cached codecs also
erase the measured scale-method difference, while direct methods retain a
small generic metadata cost:

| Operation | Explicit format | Test-local concrete | B/op | allocs/op |
|---|---:|---:|---:|---:|
| `New` | 12.1 ns/op | 11.3 ns/op | 0 | 0 |
| cached `Codec.Parse` | 9.62 ns/op | 9.62 ns/op | 0 | 0 |
| direct `Decimal.AppendTo` | 16.4 ns/op | 15.7 ns/op | 0 | 0 |

The only public format path is explicit `PriceUint*` / `AmountUint*`
composition. These formats embed concrete backend providers. A measured
alternative that selected backend operations through a generic type switch
cost roughly 21-29% on parse/format microbenchmarks, so it was not adopted.
Cached codecs remove the measured generic scale-metadata difference and are
the intended hot-path API.

## Go 1.26 review

The module requires Go 1.26.5. The upgrade was reviewed against the
[Go 1.26 release notes](https://go.dev/doc/go1.26) and measured on identical
benchmark workloads with Go 1.26.2, 1.26.3, and 1.26.5 before changing source.

Useful changes adopted here:

- Benchmarks use `testing.B.Loop`. Go 1.26 permits the loop body to inline
  while still keeping inputs and results alive, reducing benchmark-harness
  distortion.
- Production loops use the built-in `min`, `max`, and integer `range` forms
  selected by the Go 1.26 modernizer.
- Go 1.26 can place more variable-sized slice backing stores on the stack.
  Sailfish's caller-buffer hot paths were already at zero allocations, so the
  final allocation matrix remains unchanged rather than claiming an
  unobserved gain.

Changes reviewed but not adopted:

- Self-referential generic constraints do not provide associated types or
  remove the backend operations needed for `uint64` and the external
  `uint256.Int`; changing the public type model would add wrappers without a
  measured hot-path benefit.
- `new(expr)`, `errors.AsType`, and `bytes.Buffer.Peek` are not used on decimal
  parse, format, compare, or arithmetic paths.
- Green Tea GC improves allocation-heavy programs, while the measured Sailfish
  hot paths do not allocate. The one formatted `String` allocation is required
  ownership of the returned immutable string.
- Experimental architecture SIMD is unstable, amd64-only in Go 1.26, and does
  not fit this arm64-tested branchy decimal parser well enough to justify a
  production dependency.

Patch-release benchmark differences were small and inconsistent; no
toolchain-only speedup is claimed. The reason for pinning the current patch is
compiler/runtime correctness and security maintenance, plus the corrected
benchmarking contract.

## Allocation ownership

Allocation profiling attributes the formatted `String` allocation to
`Uint64Units.unitString`. It is retained by contract because the caller owns
the returned immutable string.

An early failure-path benchmark exposed `16 B/op, 1 alloc/op` from converting a
typed string constant into an `error` interface inside generic code. The final
implementation pre-boxes the fixed exported constants once; known parse,
precision, range, scale, overflow, and underflow errors now return at
`0 B/op, 0 allocs/op`.

Escape analysis identifies only these relevant classes:

- `string(out)` in `unitString`: required owned string result;
- `make([]byte, n)` in `growBy`: only when caller capacity is insufficient;
- `MarshalText` and `MarshalJSON` result slices: required owned encoding
  results;
- unknown future `Error` values may box on the cold fallback path. All exported
  fixed errors are pre-boxed once and return with zero per-call allocations.

## CPU ownership

The wide-value profile is dominated by:

- pairwise decimal chunk parsing (`parseUint64Chunk`);
- four-limb decimal multiplication/addition while parsing;
- `bits.Div64` and fixed-width chunk formatting while appending uint256 values.

The market-shape matrix separates ordinary CEX values from maximum-width
values. Runtime-scale one-limb parsing uses `Uint256Codec` and stays below
10 ns on the measured host. A raw four-limb append still performs decimal
division and takes roughly 146-152 ns; retaining canonical input or calling
`Canonical` once reduces repeated appends of that same value to about 4.4 ns.

These are the expected arithmetic owners. The current pure-Go path has no heap
traffic. `unsafe` or assembly should require a concrete production target and a
repeatable profile showing that these operations materially constrain the
system, followed by per-architecture correctness and fallback coverage.
