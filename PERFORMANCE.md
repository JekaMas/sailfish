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
