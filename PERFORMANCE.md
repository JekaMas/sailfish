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
