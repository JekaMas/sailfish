# Performance Notes

## Environment

```text
Go:    go1.26.2
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
```

## Snapshot

| Benchmark | Typical result | B/op | allocs/op |
|---|---:|---:|---:|
| `CodecParse/uint64/canonical` | 9.9-10.1 ns/op | 0 | 0 |
| `CodecParse/uint64/compact` | 9.9 ns/op | 0 | 0 |
| `CodecParse/uint64/bytes` | 9.8 ns/op | 0 | 0 |
| `CodecParse/uint64/invalid` | 10.6 ns/op | 0 | 0 |
| `CodecParse/uint256/canonical` | 52-56 ns/op | 0 | 0 |
| `CodecParse/uint256/max` | 77-84 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/parse_into_runtime_codec/cex_scale6_one_limb` | 9.4 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/cex_scale6_one_limb` | 3.5 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_retained/scale18_four_limb` | 4.4 ns/op | 0 | 0 |
| `Uint256MarketHotPaths/append_runtime_codec/cex_scale6_one_limb` | 14.2 ns/op | 0 | 0 |
| `AppendTo/uint64/retained` | 2.9 ns/op | 0 | 0 |
| `AppendTo/uint64/formatted` | 13.8 ns/op | 0 | 0 |
| `AppendTo/uint256/formatted` | 146-152 ns/op | 0 | 0 |
| `String/uint64/retained` | 2.1 ns/op | 0 | 0 |
| `String/uint64/formatted` | 30.1 ns/op | 16 | 1 |
| `Compare/uint64/same-scale` | 2.1 ns/op | 0 | 0 |
| `Compare/uint256/same-scale` | 6.4 ns/op | 0 | 0 |
| `Compare/cross-scale` | 52 ns/op | 0 | 0 |
| `AddAssign/uint64` | 4.3 ns/op | 0 | 0 |
| `AddAssign/uint256` | 13.5 ns/op | 0 | 0 |
| `ReferenceStrconvSplitUint64` | 20.7 ns/op | 0 | 0 |

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
