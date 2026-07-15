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

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench '^BenchmarkPerformanceCeilings$' \
  -benchmem -benchtime=500ms -count=15

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench '^BenchmarkOptimizationFormatReverseSWAR$' \
  -benchmem -benchtime=300ms -count=10

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench 'BenchmarkJSONHotPaths/(append|marshal_direct|marshal_go_json|unmarshal_direct|unmarshal_go_json)/' \
  -benchmem -benchtime=500ms -count=10

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench 'BenchmarkCBORRealMarketScalars' \
  -benchmem -benchtime=500ms -count=10

GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench 'BenchmarkCBORRealMarketBars' \
  -benchmem -benchtime=500ms -count=10
```

## Snapshot

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `CodecParse/uint64/canonical` | 7.76 ns/op | 0 | 0 |
| `CodecParse/uint64/compact` | 7.58 ns/op | 0 | 0 |
| `CodecParse/uint64/bytes` | 7.53 ns/op | 0 | 0 |
| `CodecParse/uint64/invalid` | 11.6 ns/op | 0 | 0 |
| `CodecParse/uint256/canonical` | 49.4 ns/op | 0 | 0 |
| `CodecParse/uint256/max` | 64.8 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/parse_runtime_codec/cex_scale6_one_limb` | 8.59 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/parse_into_runtime_codec/cex_scale6_one_limb` | 9.28 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/cex_scale6_one_limb` | 3.52 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/scale18_four_limb` | 4.22 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_runtime_codec/cex_scale6_one_limb` | 12.1 ns/op | 0 | 0 |
| `AppendTo/uint64/retained` | 2.95 ns/op | 0 | 0 |
| `AppendTo/uint64/formatted` | 9.8 ns/op | 0 | 0 |
| `AppendTo/uint256/formatted` | 113 ns/op | 0 | 0 |
| `String/uint64/retained` | 2.14 ns/op | 0 | 0 |
| `String/uint64/formatted` | 27.1 ns/op | 16 | 1 |
| `Compare/uint64/same-scale` | 2.14 ns/op | 0 | 0 |
| `Compare/uint256/same-scale` | 6.57 ns/op | 0 | 0 |
| `Compare/cross-scale` | 52.7 ns/op | 0 | 0 |
| `AddAssign/uint64` | 4.48 ns/op | 0 | 0 |
| `AddAssign/uint256` | 13.8 ns/op | 0 | 0 |
| `IntegerConversions/from_big_int/uint64` | 3.93 ns/op | 0 | 0 |
| `IntegerConversions/from_big_int/uint256` | 8.30 ns/op | 0 | 0 |
| `IntegerConversions/from_u256/uint64` | 2.90 ns/op | 0 | 0 |
| `IntegerConversions/from_u256/uint256` | 7.29 ns/op | 0 | 0 |
| `IntegerConversions/to_u256/uint64` | 2.19 ns/op | 0 | 0 |
| `IntegerConversions/to_u256/uint256` | 3.00 ns/op | 0 | 0 |
| `IntegerConversions/to_big_int_reused/uint64` | 4.21 ns/op | 0 | 0 |
| `IntegerConversions/to_big_int_reused/uint256` | 6.15 ns/op | 0 | 0 |
| `IntegerConversions/to_big_int_fresh/uint256` | 27.1 ns/op | 32 | 1 |
| `ReferenceStrconvSplitUint64` | 19.3 ns/op | 0 | 0 |

## Native reverse-SWAR formatting

The native formatter keeps the pair-table path for scale zero, below-scale
values, and measured widths where its shorter dependency chain wins. Scaled
widths 5-8 and 14-20 use arithmetic-packed eight-digit blocks. The permanent
matrix includes selected and adjacent protected widths:

| Scaled digit width | v1.0.2 | Current | Change | B/op | allocs/op |
|---:|---:|---:|---:|---:|---:|
| 5 | 8.51 ns | 7.98 ns | -6.2% | 0 | 0 |
| 7 | 8.71 ns | 7.72 ns | -11.4% | 0 | 0 |
| 8, scale 5 | 9.59 ns | 7.09 ns | -26.0% | 0 | 0 |
| 9, protected | 9.80 ns | 9.93 ns | +1.3% | 0 | 0 |
| 12, protected | 11.7 ns | 11.8 ns | +1.1% | 0 | 0 |
| 14 | 13.0 ns | 11.6 ns | -10.7% | 0 | 0 |
| 16 | 14.2 ns | 11.8 ns | -17.3% | 0 | 0 |
| 19 | 16.9 ns | 14.0 ns | -16.9% | 0 | 0 |
| 20 | 17.1 ns | 13.9 ns typical | about -18% | 0 | 0 |

The width selector is a rotated compile-time bitset. Disassembly shows one
`RORW` and one bit test on arm64, avoiding the sign and width guards emitted
for an ordinary variable shift. Explicit helper preconditions leave one
intentional bounds proof per output shape; all downstream block-slice checks
are eliminated. The common public append moved from 12.6 ns to 9.8 ns.

CPU profiles assign the selected path to `packedASCII8`,
`putPacked8WithPoint`, `fillPackedScaled64`, digit counting, and output growth.
The allocation profile contains only profiler infrastructure: caller-buffer
formatting remains `0 B/op, 0 allocs/op`.

## JSON

JSON uses a quoted fixed-scale decimal string. `AppendJSON` is the
caller-buffer API; `MarshalJSON` returns one owned result; `UnmarshalJSON`
parses canonical quoted decimals directly and delegates escaped JSON strings
to `goccy/go-json` for standards-compliant unescaping.

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `JSONHotPaths/append/native_retained` | 4.41 ns/op | 0 | 0 |
| `JSONHotPaths/append/native_formatted` | 14.4 ns/op | 0 | 0 |
| `JSONHotPaths/append/wide_retained` | 7.31 ns/op | 0 | 0 |
| `JSONHotPaths/append/wide_formatted` | 141 ns/op | 0 | 0 |
| `JSONHotPaths/marshal_direct/native_retained` | 20.1 ns/op | 16 | 1 |
| `JSONHotPaths/marshal_direct/native_formatted` | 34.9 ns/op | 16 | 1 |
| `JSONHotPaths/marshal_direct/wide_retained` | 32.4 ns/op | 96 | 1 |
| `JSONHotPaths/marshal_direct/wide_formatted` | 181 ns/op | 96 | 1 |
| `JSONHotPaths/unmarshal_direct/native_canonical` | 15.2 ns/op | 0 | 0 |
| `JSONHotPaths/unmarshal_direct/wide_canonical` | 81.0 ns/op | 0 | 0 |
| `JSONHotPaths/unmarshal_direct/native_escaped` | 121 ns/op | 40 | 2 |
| `JSONHotPaths/marshal_go_json/wide_formatted` | 506 ns/op | 240 | 3 |
| `JSONHotPaths/unmarshal_go_json/wide_canonical` | 232 ns/op | 192 | 2 |

The selected wide `MarshalJSON` reserves the maximum possible uint256 text
size instead of computing an exact length before formatting. This removes a
duplicate four-limb decimal split: ten controlled runs improved the formatted
wide path from 346 ns to 213 ns (-38.4%) while preserving its single owned
result allocation. Parse-first canonical decode improved the native path from
16.5 ns to 14.0 ns (-15.1%) at zero allocations. Escaped input is a colder
standards path and changed from 109 ns to 120 ns; its two allocations remain
owned by JSON unescaping.

CPU profiles attribute formatted wide output to reciprocal `bits.Mul64`
reductions and fixed-width chunk writing in `splitUint256Decimal`; canonical
decode is decimal parsing and scale validation. Allocation and escape profiles
attribute the direct marshal's single allocation to its returned slice.
Canonical direct decode allocates nothing. The goccy integration rows include
interface and library ownership in addition to Sailfish work.

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
| `CBORUint256Widths/runtime_codec_decode/one_limb` | 4.24 ns/op | 0 | 0 |
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
oracle uses the same 93-byte wire as deterministic fxamacker `toarray`. That
size describes one synthetic value set, not the schema's minimum or a typical
venue record.

Its one decode allocation owns the parent record's symbol string; all four
price fields and the wide amount decode without allocation through
`ParseCBORFirst`.

CPU profiles place direct CBOR work in shortest-form integer validation,
`uint256` byte-width selection, big-endian limb transfer, and scale validation.
Memory profiles contain no per-operation allocation from append or decode.
Cached `FixedDecimalCodec` methods avoid repeated generic scale metadata work and are the
recommended hot-loop surface. The figures above are means of ten
`GOMAXPROCS=1`, 200 ms runs on the documented host.

### Real-market CBOR records

`testdata/market_cbor_samples.json` is a fixed July 15, 2026 snapshot of 300
unique market identities:

| Cohort | Selection | Official data |
|---|---|---|
| MEXC spot | Top 100 online USDT markets by 24h `quoteVolume` | `exchangeInfo`, `ticker/24hr` |
| Hyperliquid spot | Top 100 main-dex markets by 24h `dayNtlVlm` | `spotMetaAndAssetCtxs`, `l2Book` |
| Hyperliquid perps | Top 100 main-dex markets by 24h `dayNtlVlm` | `metaAndAssetCtxs`, `l2Book` |

The source contracts are documented by the
[MEXC Spot API](https://mexcdevelop.github.io/apidocs/spot_v3_en/),
[Hyperliquid spot info API](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint/spot),
and [Hyperliquid perpetual info API](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint/perpetuals).
Hyperliquid spot metadata is joined by its canonical universe `index`, then
queried as `@index`; array position is not treated as market identity.

The fixture deduplicates by `(venue, market_type, symbol)`. Within each market,
it removes repeated/non-positive observations and numerically orders the
remaining values. Price cases are min, lower-third, upper-third, and max from
context/ticker/book prices. Quantity cases are metadata quantum plus book sizes
reduced to min, median, and max. It does not invent values for sparse markets:
298 markets have four prices and 297 have three quantities; the remaining
markets retain the one or two distinct observations that were available. The
four zero-volume Hyperliquid spot tail entries are retained because only 96
main-dex spot markets reported positive `dayNtlVlm` at snapshot time.

The theoretical structural floor for this exact 14-field record is 15 bytes:
one array header, one record-kind integer, one empty-text marker, five zero
decimal integers, and seven zero lifecycle integers. It is not a valid market
record because the symbol and market values are empty/zero. There is no finite
schema maximum without a symbol-length bound. Realistic snapshot records are:

| Cohort | Quantity case | Min | p50 | p95 | Max | Mean bytes |
|---|---|---:|---:|---:|---:|---:|
| MEXC spot | minimum | 55 | 60 | 68 | 75 | 60.96 |
| MEXC spot | median | 59 | 63 | 71 | 78 | 63.66 |
| MEXC spot | maximum | 62 | 64 | 74 | 78 | 65.92 |
| Hyperliquid spot | minimum | 48 | 56 | 64 | 69 | 57.42 |
| Hyperliquid spot | median | 48 | 60 | 66 | 71 | 60.18 |
| Hyperliquid spot | maximum | 48 | 60 | 68 | 73 | 60.80 |
| Hyperliquid perps | minimum | 53 | 56 | 64 | 66 | 57.86 |
| Hyperliquid perps | median | 56 | 60 | 67 | 68 | 60.61 |
| Hyperliquid perps | maximum | 57 | 60 | 68 | 69 | 61.52 |

Ten 500 ms runs per sub-benchmark produced these arithmetic means:

| Cohort | Quantity case | Encode | Decode | Encode heap | Decode heap |
|---|---|---:|---:|---:|---:|
| MEXC spot | minimum | 52.2 ns | 120.8 ns | 0 B / 0 allocs | 9 B / 1 alloc |
| MEXC spot | median | 54.3 ns | 125.3 ns | 0 B / 0 allocs | 9 B / 1 alloc |
| MEXC spot | maximum | 53.4 ns | 122.1 ns | 0 B / 0 allocs | 9 B / 1 alloc |
| Hyperliquid spot | minimum | 54.2 ns | 108.3 ns | 0 B / 0 allocs | 4 B / 1 alloc |
| Hyperliquid spot | median | 56.0 ns | 110.6 ns | 0 B / 0 allocs | 4 B / 1 alloc |
| Hyperliquid spot | maximum | 54.7 ns | 107.4 ns | 0 B / 0 allocs | 4 B / 1 alloc |
| Hyperliquid perps | minimum | 53.2 ns | 106.1 ns | 0 B / 0 allocs | 4 B / <1 alloc |
| Hyperliquid perps | median | 55.2 ns | 111.0 ns | 0 B / 0 allocs | 4 B / <1 alloc |
| Hyperliquid perps | maximum | 54.8 ns | 107.7 ns | 0 B / 0 allocs | 4 B / <1 alloc |

The fractional Hyperliquid-perp allocation count is Go's per-operation average
for short owned symbol strings and rounds to zero in benchmark output; the
reported bytes remain nonzero. FixedDecimal scalar encoding and decoding are 5.7-
7.0 ns across the six price/quantity cohorts, with `0 B/op, 0 allocs/op`.
Observed scalar wires range from 1 to 9 bytes. The benchmark reports sample
count and min/p50/p95/max/mean wire size alongside timing, and the unit test
locks all nine bar distributions so fixture drift cannot silently redefine the
result.

## Measured implementation ceilings

`BenchmarkPerformanceCeilings` pairs a public hot operation with the
narrowest same-binary kernel that performs equivalent numeric or ownership
work. These are optimization bounds for this implementation, not portable CPU
cycle guarantees.

| Operation | Public | Kernel | B/op | allocs/op |
|---|---:|---:|---:|---:|
| Runtime uint256 parse | 8.75 ns | 8.63 ns | 0 | 0 |
| Runtime parse into | 9.23 ns | parser + assignment | 0 | 0 |
| Runtime one-limb append | 12.5 ns | 11.1 ns | 0 | 0 |
| Runtime formatted length | 3.06 ns | 2.81 ns | 0 | 0 |
| Retained native append | 2.95 ns | 2.14 ns | 0 | 0 |
| Retained wide append | 5.01 ns | 3.03 ns | 0 | 0 |
| Native compare | 2.09 ns | 2.27 ns | 0 | 0 |
| Wide compare | 5.63 ns | about 3 ns | 0 | 0 |
| Runtime CBOR length | 2.60 ns | 2.68 ns | 0 | 0 |
| Runtime CBOR decode | 4.24 ns | 4.23 ns | 0 | 0 |
| Runtime CBOR decode into | 4.84 ns | 4.53 ns | 0 | 0 |

String/byte parse, caller-owned parse, raw/retained output, compare, checked
addition, CBOR append/length/decode, and first-item decode all have explicit
ceiling rows. Owned `String` and marshal operations are excluded from a
zero-allocation target because returning owned immutable storage is their
contract.

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
| `NewFixedDecimal` | 12.1 ns/op | 11.3 ns/op | 0 | 0 |
| cached `FixedDecimalCodec.Parse` | 9.62 ns/op | 9.62 ns/op | 0 | 0 |
| direct `FixedDecimal.AppendTo` | 16.4 ns/op | 15.7 ns/op | 0 | 0 |

The only public format path is explicit
`PriceInUint*Units[DecimalPlacesN]` / `AmountInUint*Units[DecimalPlacesN]`
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
- escaped `UnmarshalJSON`: JSON unescape string/interface ownership in
  `goccy/go-json`; canonical quoted decimals bypass this path;
- an unrecognized `Error` value uses direct interface conversion. Every error
  produced by the package is a pre-boxed fixed constant and returns with zero
  per-call allocations.

## CPU ownership

The wide-value profile is dominated by:

- pairwise and validated-SWAR decimal chunk parsing;
- four-limb decimal multiplication/addition while parsing;
- reciprocal `bits.Mul64` reductions and fixed-width chunk formatting while
  appending uint256 values.

The native formatting profile separates into branchless decimal digit count,
reverse-SWAR lane conversion, in-register point insertion, and exact-width
stores. No table lookup or divide-by-100 loop remains on selected widths.

The market-shape matrix separates ordinary CEX values from maximum-width
values. Runtime-scale one-limb parsing uses `Uint256FixedDecimalCodec` and stays below
10 ns on the measured host. A raw four-limb append performs reciprocal decimal
chunk reduction and takes roughly 108-113 ns; retaining canonical input or
calling `Canonical` once reduces repeated appends of that same value to about
4.4 ns.

These are the expected arithmetic owners and have no heap traffic. One
read-only unsafe load is selected at build time for amd64/arm64 SWAR words;
other architectures use the portable load. Assembly and runtime feature
dispatch still require a concrete production target and repeatable evidence
that one new implementation wins across the supported workload.

## Rational and cross-scale arithmetic

Run the public operations and their standard-library/arithmetic ceilings with:

```sh
GOMAXPROCS=1 GOWORK=off go test -run '^$' \
  -bench 'Benchmark(BigRatCeilings|CrossScaleCeilings|RationalAndCrossScaleOperations)$' \
  -benchmem -count=10
```

Representative Apple M1 Max / Go 1.26.5 results:

| Operation | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| Exact native `FromBigRat` | about 10 ns | 0 | 0 |
| Exact wide `FromBigRat` | about 31 ns | 0 | 0 |
| Integral native `ToBigRat` | about 11 ns | 0 | 0 |
| Fractional native `ToBigRat` | about 108 ns | 24 | 3 |
| Native scale-2 to scale-5 `Rescale` | about 14 ns | 0 | 0 |
| Native mixed-scale `AddAs` | about 20 ns | 0 | 0 |
| Same-scale `Denominated` add | about 4.8 ns | 0 | 0 |
| Mixed-scale `AddDenominatedAs` | about 21 ns | 0 | 0 |

`BenchmarkBigRatCeilings` reports `SetFrac`, `SetUint64`, numerator/denominator
read, and four-limb multiply separately. This prevents `math/big` allocations
from being attributed to Sailfish's allocation-free conversion kernels.
