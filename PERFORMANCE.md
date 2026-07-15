# Performance Decisions

This file records only implemented choices and measured rejections from the
current scalar optimization round. `main` contains one selected implementation
per operation; rejected prototypes and replaced implementations are removed.

Each accepted decision records:

- the measured owner and workload;
- the exact before/after benchmark artifacts;
- `ns/op`, `B/op`, and `allocs/op` changes;
- profile, escape-analysis, and bounds-check evidence;
- correctness gates;
- why the selected implementation is the package's single production path.

Rejected candidates are listed briefly so the same experiment is not repeated,
but their code is not retained.

## 2026-07-14: Dense Wide Parsing Uses Eight-Digit SWAR

**Decision:** parse dense 8-19 digit chunks inside `uint256` inputs with a
pure-Go eight-byte SWAR validator/converter. Keep the native `uint64` parser's
single pairwise implementation unchanged, including for short wide prefixes.

The baseline CPU profile attributed about 50% of parse samples to the string
and byte pairwise chunk parsers. A same-binary, `GOMAXPROCS=1`, ten-run
comparison measured these changes:

| Public parse workload | Time change | Allocations |
|---|---:|---:|
| `uint256`, 19 digits, string / bytes | -21.3% / -18.7% | 0 -> 0 |
| `uint256`, 20 digits, string / bytes | -19.9% / -19.9% | 0 -> 0 |
| `uint256`, 38 digits, string / bytes | -20.8% / -19.7% | 0 -> 0 |
| `uint256`, 57 digits, string / bytes | -16.5% / -21.0% | 0 -> 0 |
| `uint256`, 77 digits, string / bytes | -21.9% / -19.2% | 0 -> 0 |
| Native `uint64`, 1-20 digits | within +/-1.7% | 0 -> 0 |

Artifacts:

- `.codex_tmp/cpu_algorithm_round/c1_wide_specialized_same_binary.txt`
- `.codex_tmp/cpu_algorithm_round/c1_production_parse_cpu.pprof`
- `.codex_tmp/cpu_algorithm_round/c1_production_parse_mem.pprof`
- `.codex_tmp/cpu_algorithm_round/c1_production_escape_bce.txt`

Dense parsing validates every byte, accepts the same syntax, and returns the
same range errors. Tests cover lengths 8-19 and every invalid-byte position for
both `string` and `[]byte`. The production parser remains allocation-free.

An attempted single upfront bounds check before the eight-byte load was
rejected. It measured about 0.4% slower on the stable 8-byte cases and no
statistically significant long-input win. No alternate loader remains in the
package.

## 2026-07-15: Native Fixed-Scale Formatting Splits Integer And Fraction

**Decision:** when a native value has digits on both sides of the decimal
point, divide once by the fixed-scale power and format the integer and fraction
directly into their final positions. This replaces formatting into a shifted
destination followed by a prefix copy. Scale-zero and below-scale values keep
their existing single formatting paths.

A same-binary, `GOMAXPROCS=1`, twelve-run comparison measured:

| Native append workload | Time change | Allocations |
|---|---:|---:|
| 9 digits, scale 5 | -18.9% | 0 -> 0 |
| 10 digits, scale 5 | -13.5% | 0 -> 0 |
| 16 digits, scale 9 | -7.1% | 0 -> 0 |
| 19 digits, scale 18 | -9.5% | 0 -> 0 |
| Scale zero / below scale | statistically unchanged | 0 -> 0 |

The former fixed-19 writer became the single width-aware fixed writer used by
both native fractions and 19-digit wide chunks. Boundary tests cover native
digit transitions and every scale from 0 through 19; randomized native and
wide round trips remain the byte-equivalence oracle.

Artifacts:

- `.codex_tmp/cpu_algorithm_round/c2_c4_same_binary.txt`
- `.codex_tmp/cpu_algorithm_round/c4_wide_same_binary.txt`
- `.codex_tmp/cpu_algorithm_round/c4_production_format_bench.txt`
- `.codex_tmp/cpu_algorithm_round/c4_production_format_cpu.pprof`
- `.codex_tmp/cpu_algorithm_round/c4_production_format_mem.pprof`
- `.codex_tmp/cpu_algorithm_round/c4_production_escape_bce.txt`

Two alternatives were rejected and removed:

- `1e9` outer chunks regressed common short, scale-5, scale-9, and below-scale
  formatting by 6-11%, while its narrow wins stayed below 4%.
- Direct decimal-point placement across `uint256` chunks regressed multi-limb
  formatting by about 2-5%. The bounded prefix copy remains the sole wide
  formatting implementation.

## 2026-07-15: Wide Formatting Keeps Base 1e19

**Decision:** retain `1e19` as the sole decimal chunk base for `uint256`
formatting. The benchmark-only `1e9` candidate was removed.

Instrumented split tests counted the actual `bits.Div64` operations:

| Value width | Base `1e19` | Base `1e9` |
|---|---:|---:|
| 65 bits | 3 | 4 |
| 128 bits | 5 | 8 |
| 192 bits | 9 | 15 |
| 256 bits | 14 | 24 |

Across scale 0, 5, and 18 full append workloads, `1e9` was 14-56% slower on
the stable comparisons; split-only cost increased by 49-63% at 65-128 bits
and remained substantially worse at larger widths. Both variants were
allocation-free, so there was no ownership tradeoff to justify the extra
division work.

Artifacts:

- `.codex_tmp/cpu_algorithm_round/c3_same_binary.txt`
- `.codex_tmp/cpu_algorithm_round/c3_1e19.txt`
- `.codex_tmp/cpu_algorithm_round/c3_1e9.txt`

## 2026-07-15: No Small-Value Formatting Fast Path

**Decision:** do not add a separate 0-99 formatting branch or cache. The
existing digit-pair table remains the sole small-number mechanism.

The candidate accelerated scale-zero hits by 52.8%, but added 10.6% to
scale-zero misses and 5.5% to representative scale-5 formatting. A synthetic
50% hit workload improved 12.1%, but no production profile established that
hit rate. The miss regressions violate the round's gate, and retaining another
dispatch path would make general formatting workload-dependent.

Artifact: `.codex_tmp/cpu_algorithm_round/c5_same_binary.txt`.

## 2026-07-15: No General Canonical SWAR Dispatch

**Decision:** keep the scalar pairwise parser for ordinary fixed-scale
`uint64` values. The dense SWAR kernel remains restricted to wide chunks where
it has no short-path dispatch cost.

Splitting canonical integer and fraction segments into SWAR candidates
improved scale-9 and scale-18 long values by 6-10%, but regressed short and
scale-5 values by 3.6-8.1%. Representative mixed batches changed by only
0-1.6%, below the keep threshold. No branch-minimized alternate parser remains.

Artifacts:

- `.codex_tmp/cpu_algorithm_round/c6_canonical_batch_cpu.pprof`
- `.codex_tmp/cpu_algorithm_round/c6_canonical_same_binary.txt`
- `.codex_tmp/cpu_algorithm_round/c6_batch_same_binary.txt`

The measured mixed batch is about 9 ns/value on the scalar implementation.
That does not justify architecture-specific assembly, feature dispatch, or a
second implementation in this round.

## Round Result

The round retained two production changes:

1. Dense `uint256` chunks use eight-digit SWAR parsing.
2. Native fixed-scale formatting writes integer and fraction fields directly.

The final ten-run matrix kept every measured parse and pre-sized append path at
`0 B/op, 0 allocs/op`. Against the initial build, stable wide parse workloads
improved by about 17-25%, and native fixed-scale formatting improved by about
10-22% on the representative scale-5/9/18 cases. Same-binary comparisons are
the acceptance authority; the cross-build baseline contains documented host
noise on unchanged subbenchmarks.

Final artifacts:

- `.codex_tmp/cpu_algorithm_round/final_bench.txt`
- `.codex_tmp/cpu_algorithm_round/final_benchstat.txt`
- `.codex_tmp/cpu_algorithm_round/final_cpu.pprof`
- `.codex_tmp/cpu_algorithm_round/final_cpu_top.txt`
- `.codex_tmp/cpu_algorithm_round/final_mem.pprof`
- `.codex_tmp/cpu_algorithm_round/final_mem_top.txt`
- `.codex_tmp/cpu_algorithm_round/final_escape_bce.txt`

No unsafe code, assembly, architecture dispatch, compatibility path, legacy
implementation, normalization fallback, or alternate runtime algorithm was
added. `main` contains one selected implementation for each length/width
specialization.

## 2026-07-15: v1 Runtime Codec Reaches Its Measured Ceilings

**Decision:** store `Uint256Codec` scale directly in its single byte and omit
scale decoding from CBOR operations, whose wire contract contains units only.
The zero value remains a valid scale-0 codec, the type remains one byte, and
no public or wire semantics change.

A measured implementation ceiling is the narrowest same-binary kernel that
performs equivalent numeric and ownership work. It is deliberately not a
cycle-count estimate. Fifteen `GOMAXPROCS=1`, 500 ms runs produced:

| Public hot operation | Public | Equivalent kernel | Heap |
|---|---:|---:|---:|
| Runtime string parse, scale 6 | 8.75 ns | 8.63 ns complete parser | 0 B / 0 allocs |
| Runtime parse into caller storage | 9.23 ns | 8.63 ns parser plus assignment | 0 B / 0 allocs |
| Raw one-limb append, scale 6 | 12.5 ns | 11.1 ns complete formatter | 0 B / 0 allocs |
| Runtime formatted length | 3.06 ns | 2.81 ns complete length path | 0 B / 0 allocs |
| Retained native append | 2.95 ns | 2.14 ns direct immutable-text copy | 0 B / 0 allocs |
| Retained wide append | 5.01 ns | 3.03 ns direct immutable-text copy | 0 B / 0 allocs |
| Native same-format compare | 2.09 ns | 2.27 ns direct integer compare | 0 B / 0 allocs |
| Wide same-format compare | 5.63 ns | about 3 ns direct limb compare | 0 B / 0 allocs |
| Runtime CBOR length | 2.60 ns | 2.68 ns complete uint256 length | 0 B / 0 allocs |
| Runtime CBOR decode | 4.24 ns | 4.23 ns complete uint256 decoder | 0 B / 0 allocs |
| Runtime CBOR decode into | 4.84 ns | 4.53 ns complete decode and assignment | 0 B / 0 allocs |

Compared with the pre-round runtime wrapper, string parse improved from
9.80 ns to 8.75 ns (-10.7%, `p=0.000`) and CBOR decode improved from 6.74 ns
to 4.24 ns (-37.1%, `p=0.000`). The complete parse and CBOR kernels were
unchanged. The runtime-scale append comparison was statistically noisy; its
stable isolated result remains about 12-13 ns with zero allocations, so no
formatting speedup is claimed from this round.

Rejected experiments were removed:

- A value-only parse helper benchmarked faster when called directly, but made
  `Uint256Codec.Parse` exceed the compiler's inlining budget and did not
  improve the public method.
- A generic type-switch compare regressed wide compare from about 5.5 ns to
  about 8.4 ns.
- Provider-dispatched checked addition improved some wide-only samples but did
  not improve native and wide backends consistently; the single closed-unit
  type switch remains production code.
- Raw formatting, retained output, and first-item CBOR decode keep their
  existing algorithms. Their residual cost is output construction, suffix
  ownership, or generic method dispatch; no second implementation is retained.

Artifacts:

- `.codex_tmp/v1_ceiling/baseline_ceilings.txt`
- `.codex_tmp/v1_ceiling/final_ceilings.txt`
- `.codex_tmp/v1_ceiling/final_vs_baseline.txt`
- `.codex_tmp/v1_ceiling/final_escape.txt`
- `.codex_tmp/v1_ceiling/final_cpu.pprof`
- `.codex_tmp/v1_ceiling/final_mem.pprof`

The v1 release therefore keeps pure Go, one implementation per operation,
strict constructor-only canonicalization, and zero allocation across all
caller-buffer runtime codec hot paths.

## 2026-07-15: Real-Market CBOR Size Uses Value Distributions

**Decision:** retain the 93-byte positional bar as a byte-equivalence oracle,
not as a storage-size estimate. Measure aggregate CBOR with deduplicated,
positive observations from 100 MEXC spot, 100 Hyperliquid spot, and 100
Hyperliquid perpetual markets.

The fixed snapshot contains each `(venue, market_type, symbol)` once. It keeps
distinct observed min/quantile/max prices and quantities at each market's
venue scale. Three deliberate bar workloads select minimum, median, or maximum
quantity; they share the same market set instead of duplicating fixture rows.

The exact 14-field schema has a 15-byte empty/zero structural floor and no
finite maximum without a symbol-length bound. The 900 realistic bar cases
measure 48-78 bytes, with cohort/mode means from 57.42 to 65.92 bytes. Ten-run
encoding means are 52.2-56.0 ns at zero allocations. Decode means are
106.1-125.3 ns; its only ownership cost is the enclosing record's symbol
string. Decimal scalar encode/decode remains allocation-free.

This adds benchmark and test evidence only. No production codec, wire format,
normalization, compatibility decoder, fallback, or alternate implementation
was introduced.

## 2026-07-15: JSON Uses Single-Pass Wide Marshal And Parse-First Decode

**Decision:** keep manual quoted-decimal `AppendJSON`, reserve the bounded
maximum output for unretained `uint256.Int` values in `MarshalJSON`, and parse
ordinary quoted decimal bytes before scanning for JSON escapes. Escaped JSON
strings continue through `goccy/go-json`; no second decimal implementation is
retained.

The former wide marshal computed `Len` and then formatted. Both operations
split a multi-limb uint256 into decimal chunks, so an unretained wide value did
the expensive work twice. Reserving 81 bytes (maximum decimal text plus two
quotes) preserves the one required result allocation and formats once.

Ten `GOMAXPROCS=1`, 500 ms runs measured:

| Operation | Before | After | Heap |
|---|---:|---:|---:|
| Wide formatted `MarshalJSON` | 346 ns | 213 ns (-38.4%) | 96 B / 1 alloc |
| Wide retained `MarshalJSON` | 34.6 ns | 33.0 ns (-4.6%) | 96 B / 1 alloc |
| Native canonical `UnmarshalJSON` | 16.5 ns | 14.0 ns (-15.1%) | 0 B / 0 allocs |
| Wide canonical `UnmarshalJSON` | 83.3 ns | 82.9 ns (-0.5%) | 0 B / 0 allocs |
| Native escaped `UnmarshalJSON` | 109 ns | 120 ns (+10.0%) | 40 B / 2 allocs |

The escaped-path regression is accepted because ordinary trading API payloads
contain unescaped ASCII digits and a decimal point, while escaped JSON must
still use a complete standards decoder. Receiver-preservation tests cover
invalid decimals, malformed escapes, non-string JSON, and truncated input.
Allocation tests enforce zero-allocation caller-buffer append and canonical
decode, plus exactly one allocation for each owned marshal result.

Rejected benchmark-only alternatives were removed:

- stack-format followed by copy performed two allocations and copied the
  result;
- generic `go-json` marshal/unmarshal remained substantially slower and more
  allocation-heavy than the direct methods;
- a separate manual JSON unescaper was not introduced because it would
  duplicate standards behavior for a cold path.

Artifacts:

- `.codex_tmp/json_optimization/baseline_corrected.txt`
- `.codex_tmp/json_optimization/production_after.txt`
- `.codex_tmp/json_optimization/production_after_benchstat.txt`
- `.codex_tmp/json_optimization/final/json_cpu.pprof`
- `.codex_tmp/json_optimization/final/json_mem.pprof`
- `.codex_tmp/json_optimization/final/json_escape.txt`

No JSON syntax, decimal normalization, canonicalization, scale, or numeric
semantics changed. There is one current JSON implementation, no compatibility
codec, no legacy decoder, and no normalization fallback.
