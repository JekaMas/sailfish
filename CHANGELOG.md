# Changelog

## v1.0.4 - 2026-07-15

- Replaced ambiguous public decimal names with one explicit API that states
  semantic kind, scaled-integer representation, and fractional decimal places:
  `FixedDecimal`, `FixedDecimalCodec`, `DecimalPlacesN`,
  `PriceInUintNUnits`, and `AmountInUintNUnits`.
- Renamed constructors, runtime codecs, format accessors, and unsupported
  decimal-place errors to describe their exact fixed-decimal contract.
- Removed the former public names entirely. No aliases, deprecated wrappers,
  compatibility layer, legacy path, or fallback remains.

## v1.0.3 - 2026-07-15

- Added selective reverse-SWAR native formatting for measured 5-8 and 14-20
  digit scaled values, improving the common caller-buffer append by about 22%
  with zero allocations.
- Inserted decimal points in packed registers and retained exact-width stores;
  caller-owned capacity beyond the returned slice is never modified.
- Added permanent per-width benchmarks, randomized packed-digit checks,
  caller-tail ownership tests, CPU/allocation profiles, BCE diagnostics, and
  assembly-backed width-selector evidence.
- Kept pair-table formatting for rejected widths and retained base `1e19` for
  `uint256`; no alternate formatter, compatibility path, or fallback remains.

## v1.0.2 - 2026-07-15

- Added contiguous raw-unit batch APIs for cache-efficient numeric scans.
- Replaced the native decimal-width comparison tree with a measured
  branchless bit-length and borrow-bit correction.
- Kept the existing CBOR switch and SWAR decimal parser after branchless and
  `segmentio/asm` candidates measured slower or failed the exact grammar gate.
- Extended release validation with decimal-width distributions, boundary and
  randomized reference tests, assembly inspection, and Linux `perf stat`
  instructions for PMU-enabled hosts.

## v1.0.0 - 2026-07-15

- Strict unsigned fixed-scale decimals backed by `uint8`, `uint16`, `uint32`,
  `uint64`, or `uint256.Int`.
- Distinct price and amount formats with compile-time fractional scale.
- Runtime-scale `Uint256FixedDecimalCodec` for trusted venue metadata.
- Caller-buffer text, JSON, and preferred deterministic CBOR APIs.
- Checked same-format arithmetic and cross-format comparison.
- Zero-allocation parse, append, compare, arithmetic, and direct CBOR hot
  paths when caller storage has sufficient capacity.
- One current implementation and one canonical wire format; no compatibility,
  legacy, fallback, or numbered codec paths.
