# Performance Decisions

This file records only implemented choices and measured rejections from the
current scalar optimization round. `main` contains one selected implementation
per operation; rejected prototypes and replaced implementations are removed.

Each accepted decision records:

- the measured owner and workload;
- the exact before/after benchmark results;
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

Two alternatives were rejected and removed:

- `1e9` outer chunks regressed common short, scale-5, scale-9, and below-scale
  formatting by 6-11%, while its narrow wins stayed below 4%.
- Direct decimal-point placement across `uint256` chunks regressed multi-limb
  formatting by about 2-5%. The bounded prefix copy remains the sole wide
  formatting implementation.

## 2026-07-15: Selected Native Widths Use Reverse SWAR

**Decision:** use arithmetic-packed eight-digit conversion for scaled native
values with 5-8 or 14-20 digits. Keep the pair-table integer/fraction writer
for every other width. The public API, one-scaled-integer representation, and
canonical text remain unchanged.

The selected kernel converts two four-digit lanes, then four two-digit lanes,
in parallel inside one `uint64`. Reciprocal multiply/shift arithmetic replaces
the serial divide-by-100 chain. For 5-8 digits the decimal point is inserted in
the packed word before exact-width stores. For 14-20 digits, independent base
`1e8` blocks are written with one bounded prefix move. A rotated width bitset
selects the measured widths with one `RORW` and bit test on arm64; an ordinary
variable shift was rejected because Go emitted sign and shift-width guards.

Same-host, same-toolchain comparisons against `v1.0.2` measured:

| Complete formatting workload | Time change | Allocations |
|---|---:|---:|
| Public `AppendTo`, 8 digits / scale 5 | -22.2% | 0 -> 0 |
| Public newly formatted `String` | -9.4% | 1 -> 1 |
| 5 / 6 / 7 scaled digits | -6.2% / -5.3% / -11.4% | 0 -> 0 |
| 8 scaled digits, scales 1 / 5 / 7 | -28.8% / -26.0% / -23.7% | 0 -> 0 |
| 14 / 15 / 16 scaled digits | -10.7% / -10.0% / -17.3% | 0 -> 0 |
| 17 / 18 / 19 / 20 scaled digits | about -5% / -9% / -17% / -18% | 0 -> 0 |
| Rejected adjacent widths 2-4 and 9-13 | within 0-1.3% dispatch cost | 0 -> 0 |

The formatter retains exact caller-buffer ownership. A faster eight-byte
overstore for short tails was rejected because it mutated capacity beyond the
returned slice. Tests lock this contract and compare 100,000 packed values
against `strconv`, all scale/digit boundaries against `math/big`, and randomized
format/parse round trips.

Measured alternatives were removed:

- Separate integer/fraction storage did not improve the complete scale-5 path
  and would duplicate numeric state.
- A base-`1e16` `uint256` formatter was slightly faster at two limbs but 9-23%
  slower at three and four limbs; base `1e19` remains the sole wide format.
- Raw decimal text insertion lost on protected 9-11 digit values.
- AVX-512 IFMA/VBMI cannot be executed or validated on the arm64 release host,
  and its call/dispatch contract is unjustified for a 7-14 ns scalar kernel.

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

## 2026-07-15: No Small-Value Formatting Fast Path

**Decision:** do not add a separate 0-99 formatting branch or cache. The
existing digit-pair table remains the sole small-number mechanism.

The candidate accelerated scale-zero hits by 52.8%, but added 10.6% to
scale-zero misses and 5.5% to representative scale-5 formatting. A synthetic
50% hit workload improved 12.1%, but no production profile established that
hit rate. The miss regressions violate the round's gate, and retaining another
dispatch path would make general formatting workload-dependent.

## 2026-07-15: No General Canonical SWAR Dispatch

**Decision:** keep the scalar pairwise parser for ordinary fixed-scale
`uint64` values. The dense SWAR kernel remains restricted to wide chunks where
it has no short-path dispatch cost.

Splitting canonical integer and fraction segments into SWAR candidates
improved scale-9 and scale-18 long values by 6-10%, but regressed short and
scale-5 values by 3.6-8.1%. Representative mixed batches changed by only
0-1.6%, below the keep threshold. No branch-minimized alternate parser remains.

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

That round added no unsafe code, assembly, architecture dispatch,
compatibility path, legacy implementation, normalization fallback, or
alternate runtime algorithm. The later short-token round below replaced the
portable byte-shift load on amd64/arm64 after measuring the native load.

## 2026-07-15: Shape-Specialized SWAR And Reciprocal Formatting

**Decision:** keep one hybrid scalar implementation selected by token shape:

1. pairwise accumulation for short and irregular native inputs;
2. one or two explicit SWAR8 blocks for dense wide chunks;
3. a known-point SWAR compaction for exactly 8 or 16 fixed-scale digits;
4. reciprocal 2-by-1 division for the fixed `1e19` wide-format chunk base.

On amd64 and arm64, SWAR words are loaded with one read-only unaligned native
load. Other architectures compile the endian-independent byte-shift loader.
The unsafe pointer is internal, never retained, never used for a write, and is
only reached after callers prove an eight-byte range. Cross-builds, race, and
`checkptr=2` are release gates for this architecture split.

The ten-run `GOMAXPROCS=1` comparison against `fc2777a` measured:

| Workload | Before | After | Change | Allocations |
|---|---:|---:|---:|---:|
| Dense `uint256`, 19 digits, string / bytes | 11.4 / 11.5 ns | 9.5 / 9.3 ns | -16.6% / -18.9% | 0 -> 0 |
| Dense `uint256`, 38 digits, string / bytes | 21.8 / 22.0 ns | 19.1 / 18.3 ns | -12.6% / -17.0% | 0 -> 0 |
| Dense `uint256`, maximum 78 digits, string / bytes | 47.9 / 48.4 ns | 43.6 / 43.3 ns | -9.0% / -10.6% | 0 -> 0 |
| Wide format, two limbs | 48.3 ns | 43.8 ns | -9.3% | 0 -> 0 |
| Wide format, three limbs | 81.9 ns | 70.1 ns | -14.4% | 0 -> 0 |
| Wide format, four limbs, scale 18 | 136 ns | 108 ns | -21.2% | 0 -> 0 |
| Wide format, maximum value | 171 ns | 129 ns | -24.4% | 0 -> 0 |

The fixed-point shape matrix measured 4.82/4.66 ns for an 8-digit
string/byte value, 7.20/7.15 ns for 16 digits at scale 5, and 6.47/6.22 ns for
16 digits at scale 9. The public cached-codec parse of `123.31232` improved
from about 10.1 ns to 7.76 ns. Longer canonical forms outside those shapes
changed between -2.7% and +2.7%; the accepted dispatch is therefore a
short-token latency choice, not a claim that every length improves. The cold
invalid-token benchmark moved from 10.6 ns to 11.6 ns because exact-shape
inputs now reach full-word validation before returning `ErrSyntax`.

### Algorithm choices

- **Two independent SWAR8 blocks:** explicit blocks remove the loop branch and
  expose instruction-level parallelism. The direct dense kernels improved
  about 13-16% before integration.
- **Native read:** one little-endian unaligned load beat eight shifts on the
  measured arm64 host. Build tags restrict it to amd64/arm64; the portable
  implementation remains the implementation for all other architectures.
- **Known decimal point:** the scale proves the point position. Compacting it
  with shifts and reducing the resulting word validates each digit once.
- **Invariant `1e19` division:** reciprocal `0xd83c94fb6d2ac34a` implements
  Algorithm 4 from "Improved division by invariant integers" with at most two
  corrections. A 50,000-value property test compares it with `math/big`, and
  direct boundary tests compare each 2-by-1 reduction with `bits.Div64`.

### Measured rejections

- Base-`1e8` native formatting regressed all measured widths by roughly 3-16%;
  the digit-pair writer remains the only native formatter.
- Merely specializing the divisor argument to `1e19` improved isolated calls
  slightly but did not improve complete formatting. The actual reciprocal
  algorithm above was retained because it improved the complete path.
- Generated scale-5 masks reached 2.46 ns for one direct eight-digit kernel,
  but require duplicated code by scale and token length. A shared dispatch was
  slower and the retained runtime-shift implementation covers every scale in
  one audited path.
- A shared constant-mask switch improved selected shapes modestly but remained
  slower than the retained runtime-point compaction after public dispatch.
- SSSE3/AVX2 cannot be measured on the arm64 release host. NEON, padded-input,
  and batch assembly need a real exchange-decoder workload that amortizes the
  call and feature-dispatch cost. No unmeasured assembly or new padded/batch API
  was added.
- Splitting numeric and cached decimal types changes the package's ownership
  and memory API; it is not a parser optimization and was not attempted here.

## 2026-07-15: v1 Runtime FixedDecimalCodec Reaches Its Measured Ceilings

**Decision:** store `Uint256FixedDecimalCodec` scale directly in its single byte and omit
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
  `Uint256FixedDecimalCodec.Parse` exceed the compiler's inlining budget and did not
  improve the public method.
- A generic type-switch compare regressed wide compare from about 5.5 ns to
  about 8.4 ns.
- Provider-dispatched checked addition improved some wide-only samples but did
  not improve native and wide backends consistently; the single closed-unit
  type switch remains production code.
- Raw formatting, retained output, and first-item CBOR decode keep their
  existing algorithms. Their residual cost is output construction, suffix
  ownership, or generic method dispatch; no second implementation is retained.

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
fractional decimal places. Three deliberate bar workloads select minimum,
median, or maximum
quantity; they share the same market set instead of duplicating fixture rows.

The exact 14-field schema has a 15-byte empty/zero structural floor and no
finite maximum without a symbol-length bound. The 900 realistic bar cases
measure 48-78 bytes, with cohort/mode means from 57.42 to 65.92 bytes. Ten-run
encoding means are 52.2-56.0 ns at zero allocations. Decode means are
106.1-125.3 ns; its only ownership cost is the enclosing record's symbol
string. FixedDecimal scalar encode/decode remains allocation-free.

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

No JSON syntax, decimal normalization, canonicalization, scale, or numeric
semantics changed. There is one current JSON implementation, no compatibility
codec, no legacy decoder, and no normalization fallback.

## 2026-07-15: Cache Locality And Raw Unit Batches

**Decision:** keep the minimum 24-byte native and 48-byte uint256 `FixedDecimal`
layouts, and use contiguous raw unit arrays for numeric-only batch kernels.
Add `FixedDecimalCodec.ParseUnits`, `ParseUnitsBytes`, `AppendUnits`, and `UnitsLen` so the
cache-efficient representation uses the same strict typed codec at its text
boundary. Do not add padding, pointer-indirected cold state, or a second decimal
value type.

The release host is an Apple M1 Max with a 128-byte cache line. macOS reports
128 KiB performance-core L1 data caches and 12 MiB L2 caches shared by four
performance cores. Eight `GOMAXPROCS=1`, 300 ms runs measured:

| 699,050-value random scan | Raw units | `FixedDecimal` | Raw-unit gain | Heap |
|---|---:|---:|---:|---:|
| `uint64` | 1.77 ms | 2.84 ms | 1.60x | 0 B / 0 allocs |
| `uint256.Int` | 3.15 ms | 11.7 ms | 3.71x | 0 B / 0 allocs |

Sequential scans from 2,048 through 699,050 values remained approximately
2.1 ns/value for both representations, because hardware prefetch and the
simple summation loop hide the larger stride. Random access removes that
advantage and makes the 3x native and 1.5x wide working-set expansion visible.

Fifteen `GOMAXPROCS=1`, 500 ms runs measured the raw boundary methods against
the former public composition and the same internal kernels:

| Operation | Former public composition | Raw-unit method | Kernel | Heap |
|---|---:|---:|---:|---:|
| Parse native units | 7.51 ns | 7.29 ns | 4.88 ns | 0 B / 0 allocs |
| Append native units | 13.6 ns | 12.3 ns | 10.0 ns | 0 B / 0 allocs |

The remaining public-to-kernel difference is generic scale/backend dispatch,
not allocation or representation copying. A specialized second native codec
was rejected because the measured append gain does not justify duplicating the
typed `FixedDecimalCodec` API, and raw batch locality is already achieved by the unit array.

Hardware cache hit/miss totals are deliberately not claimed. The local
`xctrace` CPU Counters template exported an empty `pmc-events` selection, and
the privileged Linux Docker VM exposed only software, tracepoint, and probe
event sources. The table is therefore working-set crossover evidence, not a
hardware miss-rate measurement.

## 2026-07-15: PGO Is Owned By The Consuming Executable

**Decision:** do not commit a Sailfish `default.pgo`. Sailfish is a library,
while Go PGO optimizes the complete executable and selects `default.pgo` from
the main package. A useful profile must therefore come from the consuming
trading application and include its actual call frequencies, dependencies,
and success/error distribution.

The package experiment used Go 1.26.5 and a 40.42-second mixed CPU profile
covering native and uint256 parsing, retained and formatted output, scalar
CBOR, 900 real-market positional CBOR bar cases, and direct/generic JSON. Ten
200 ms runs compared identical test binaries built with `-pgo=off` and a
representative benchmark profile supplied through `-pgo=<profile>`.

| Protected operation | PGO off | PGO | Delta |
|---|---:|---:|---:|
| Canonical native parse | 7.76 ns | 6.60 ns | -14.9% |
| Canonical uint256 parse | 49.2 ns | 38.8 ns | -21.2% |
| Runtime-codec one-limb uint256 parse | 8.73 ns | 8.29 ns | -5.1% |
| Four-limb uint256 append | 113 ns | 109 ns | -3.1% |
| Uint256 CBOR decode | 12.8 ns | 9.8 ns | -23.4% |
| Native CBOR decode | 8.06 ns | 8.70 ns | +7.9% |
| Canonical native JSON decode | 14.2 ns | 15.5 ns | +9.1% |
| Canonical wide JSON decode | 78.1 ns | 67.4 ns | -13.7% |

Allocations and wire sizes were unchanged. Compiler PGO diagnostics show the
mechanism rather than an inferred explanation: hot budgets permitted inlining
of the native and uint256 parse chains, CBOR append/decode helpers, fixed-digit
writers, and wide division/formatting helpers. The generated test binary grew
from 9,118,978 to 9,218,722 bytes (+1.09%). A warm-cache local build measured
0.29 seconds without PGO and 0.64 seconds with PGO; those single build timings
are orientation only, not a release gate.

This profile is rejected as a package default despite broad wins because it is
benchmark-trained, does not represent a deployed application, and regresses
two protected narrow decode paths. A consuming executable should collect
equal-duration production CPU profiles, merge them with `go tool pprof
-proto`, compare `-pgo=off` with `-pgo=<profile>` under the same traffic, and
accept only after its own latency/CPU and minority-path regression gates pass.

## 2026-07-15: Branchless FixedDecimal Width, Not Branchless Everywhere

**Decision:** replace the `decimalDigits64` comparison tree with a binary
magnitude estimate and one branchless decimal-threshold correction. Keep the
existing CBOR-length switch and short-token SWAR kernels. Do not add
`segmentio/asm`.

The accepted kernel computes an initial power with
`bits.Len64(v|1) * 1233 >> 12`, then uses the borrow bit from `bits.Sub64` to
correct that estimate around powers of ten. The `v|1` maps zero to one digit
without a separate branch. Boundary tests cover zero, `MaxUint64`, and the
values immediately below, at, and above every representable power of ten;
100,000 random values compare the result with `strconv.FormatUint`.

Ten same-binary runs measured the removed comparison tree against the accepted
kernel:

| Input distribution | Comparison tree | Bit-length correction | Delta | Heap |
|---|---:|---:|---:|---:|
| Predictable eight digits | 3.29 ns | 2.33 ns | -29.2% | 0 B / 0 allocs |
| Mixed decimal widths | 2.80 ns | 2.60 ns | -7.1% | 0 B / 0 allocs |
| Market-shaped widths | 3.11 ns | 2.34 ns | -24.8% | 0 B / 0 allocs |

The isolated result survives the public boundary. Thirty paired,
order-alternated runs of `AppendTo/uint64/formatted` improved from 12.9 ns to
12.6 ns (-1.98%, p=0.000), with 0 B/op and 0 allocs/op. The full native-width
formatting matrix was statistically neutral, and four-limb wide formatting
remained approximately 114 ns.

Generated darwin/arm64 assembly contains `CLZ`, `MUL`, `LSR`, the power-of-ten
table load, `SUBS`, and `NGC`; the decimal-width calculation has no
data-dependent branch. This is machine-code evidence for the transformation,
not a claim about hardware branch-miss counts.

Rejected alternatives:

- A branchless CBOR unsigned-integer length candidate performed four threshold
  operations unconditionally. It measured 2.49-2.50 ns across predictable and
  mixed inputs, versus 2.18-2.21 ns for the existing switch. The predictable
  branches and lower instruction count win.
- `segmentio/asm/ascii.ValidString` measured approximately 2.60 ns for 8 bytes,
  3.58 ns for 16 bytes, and 4.18 ns for 32 bytes, all allocation-free. It is
  still unsuitable here: ASCII validity is weaker than decimal grammar and
  range validation, using it would add a second pass, and its ASCII assembly
  implementation is amd64-only while darwin/arm64 selects generic Go. The
  existing Sailfish SWAR loader validates digits while accumulating their
  numeric value.

Hardware counters remain unavailable in this environment. Darwin does not
provide Linux `perf`, and the privileged Docker VM exposes only software,
tracepoint, and probe event sources, not a CPU PMU. A bare-metal or PMU-enabled
Linux follow-up should run the same benchmark binary under:

```sh
taskset -c 2 perf stat -x, -r 10 \
  -e task-clock,cycles,instructions,branches,branch-misses,cache-references,cache-misses \
  ./sailfish.test -test.run '^$' \
  -test.bench '^BenchmarkDecimalDigitsDistributions$' \
  -test.benchtime=3s -test.count=1 -test.cpu=1
```
