# Changelog

## v1.2.0 - 2026-07-15

- Added exact, fail-closed conversion to and from non-negative `math/big.Rat`.
  Exact input conversion is allocation-free; whole-integer output bypasses
  rational normalization, while fractional output documents the unavoidable
  `math/big.Rat.SetFrac` ownership/allocation floor.
- Added exact `Rescale`, `AddAs`, and `SubAs` operations across fractional
  decimal places and unit backends. Native results use a measured native
  checked kernel; wide results retain four-limb arithmetic.
- Added `Denominated` for attaching comparable token/market identity without
  inferring scale or taking over metadata, serialization, or lifecycle
  ownership. Same-format and cross-format arithmetic reject identity mismatch.
- Added randomized `big.Int`/`big.Rat` reference tests, fuzz targets,
  allocation assertions, layout checks, public examples, CPU/allocation
  profiles, escape/BCE evidence, and component ceilings.
- Rejected a provider-method max-scale candidate after it regressed rescale
  and mixed-scale add. No compatibility API, legacy path, fallback, unsafe
  rational access, pooling, or assembly was added.

## v1.1.0 - 2026-07-15

- Added direct, range-checked conversion between `FixedDecimal` and
  non-negative `big.Int` / `uint256.Int` already-scaled units.
- Added caller-owned `ToBigInt` output so repeated wide conversion remains
  allocation-free; `ToU256` returns one inline four-limb value.
- Added cross-width, ownership, allocation, property, fuzz, external API, and
  performance-ceiling coverage. Wide public constructors remain below 10 ns
  on the documented Apple M1 Max baseline.
- Documented the existing checked `Add`/`Sub` and explicit overflow/underflow
  arithmetic API; no second arithmetic contract was introduced.

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
