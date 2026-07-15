# Changelog

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
