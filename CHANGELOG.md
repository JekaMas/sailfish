# Changelog

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
- Runtime-scale `Uint256Codec` for trusted venue metadata.
- Caller-buffer text, JSON, and preferred deterministic CBOR APIs.
- Checked same-format arithmetic and cross-format comparison.
- Zero-allocation parse, append, compare, arithmetic, and direct CBOR hot
  paths when caller storage has sufficient capacity.
- One current implementation and one canonical wire format; no compatibility,
  legacy, fallback, or numbered codec paths.
