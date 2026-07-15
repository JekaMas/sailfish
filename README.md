# sailfish

`sailfish` is a fast, unsigned, fixed-decimal package for trading and
financial protocols that exchange exact values as strings such as
`"123.31232"`.

The numeric state is one scaled integer:

```text
value = units / 10^fractionalDecimalPlaces
```

Supported unit backends are `uint8`, `uint16`, `uint32`, `uint64`, and
`uint256.Int`. The common parse, append, compare, and arithmetic paths perform
no heap allocations.

Sailfish requires Go 1.26.5 or newer.

The current release is `v1.0.4`. On the documented Apple M1 Max / Go
1.26.5 benchmark host, common native formatting is 9.8 ns, runtime-scale
uint256 parsing is 8.69 ns, and direct uint256 CBOR decode is 4.20 ns. These
caller-buffer operations track measured implementation kernels and perform no
heap allocations. See
[BENCHMARKS.md](BENCHMARKS.md) and [PERFORMANCE.md](PERFORMANCE.md) for the
complete matrix and rejected alternatives.

## Single-format policy

`main` contains one current implementation and one canonical wire format. It
does not retain compatibility codecs, legacy decoders, alternate encodings, or
compatibility fallback implementations. `FixedDecimal` CBOR is always the
preferred shortest unsigned integer representation, using tag 2 only when a
`uint256.Int` does not fit in `uint64`. Input in any other representation is
rejected instead of being normalized or decoded by an older path.

Optimizations replace the previous implementation after benchmarks and the
complete correctness suite pass. They do not add parallel numbered codec
versions.

## Quick start

Select semantic kind, integer representation, and fractional decimal places
explicitly:

```go
codec, err := sailfish.NewFixedDecimalCodec[sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]]()
if err != nil {
	return err
}

price, err := codec.Parse("123.31232")
if err != nil {
	return err
}

delta, err := codec.Parse("0.00001")
if err != nil {
	return err
}

if overflow := price.AddAssign(delta); overflow {
	return sailfish.ErrOverflow
}

request := make([]byte, 0, 32)
request = codec.AppendTo(request, price)
// request == "123.31233"
```

The type name is the storage contract:

```text
PriceInUint64Units[DecimalPlaces5]
│            │         └─ exactly 5 digits after the decimal point
│            └─────────── raw numeric state is one uint64 scaled integer
└──────────────────────── semantic kind is price

numeric value = raw units / 100000
12_331_232 raw units = 123.31232
```

For `PriceInUint64Units[DecimalPlaces9]`, the maximum representable value is
`18446744073.709551615`.

## Choosing a type

Choose semantic kind, fractional decimal places, and integer capacity
independently.
The format type carries all three choices, so a price cannot be passed where
an amount is required even when both use the same decimal places and backend.

| Typical value | Suggested format | Why |
|---|---|---|
| Small bounded ratio or rate | `PriceInUint16Units[DecimalPlaces4]` | Compact units with four exact fractional digits |
| CEX price or quantity | `PriceInUint64Units[DecimalPlaces5]`, `AmountInUint64Units[DecimalPlaces8]` | Native arithmetic with explicit venue precision |
| Token amount | `AmountInUint256Units[DecimalPlaces18]` | Full EVM-width scaled units |
| Runtime venue metadata | `Uint256FixedDecimalCodec` | Fractional decimal places are validated once without storing them per value |

Use the narrowest backend whose complete scaled-integer range covers the
protocol contract. Fractional decimal places alone do not determine the
backend.

## Construction patterns

Parse canonical venue text with a cached codec on repeated paths:

```go
type PriceFormat = sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]
type Price = sailfish.FixedDecimal[PriceFormat, uint64]

priceCodec, err := sailfish.NewFixedDecimalCodec[PriceFormat]()
if err != nil {
	return err
}
price, err := priceCodec.Parse("123.31232")
if err != nil {
	return err
}
```

Construct directly from already-scaled protocol units without a text
round-trip:

```go
type AmountFormat = sailfish.AmountInUint32Units[sailfish.DecimalPlaces6]

amount, err := sailfish.NewFixedDecimalFromUnits[AmountFormat](uint32(1_234_567))
if err != nil {
	return err
}
// amount.String() == "1.234567"
```

Use distinct formats for domain boundaries:

```go
type CEXPrice = sailfish.FixedDecimal[
	sailfish.PriceInUint64Units[sailfish.DecimalPlaces5],
	uint64,
]
type TokenAmount = sailfish.FixedDecimal[
	sailfish.AmountInUint256Units[sailfish.DecimalPlaces18],
	uint256.Int,
]
```

For fractional decimal places supplied by trusted venue metadata, validate
them once and parse into caller-owned storage:

```go
codec, err := sailfish.NewUint256FixedDecimalCodec(18)
if err != nil {
	return err
}
var units uint256.Int
if err := codec.ParseInto("1.250000000000000000", &units); err != "" {
	return err
}
```

For large numeric batches, keep the raw scaled units contiguous and use the
same typed codec at the text boundary. This avoids carrying each `FixedDecimal`'s
optional retained-string header through a price or amount kernel:

```go
type PriceFormat = sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]

codec, err := sailfish.NewFixedDecimalCodec[PriceFormat]()
if err != nil {
	return err
}
units, parseErr := codec.ParseUnits("123.31232")
if parseErr != "" {
	return parseErr
}
prices := []uint64{units}

var wire [32]byte
encoded := codec.AppendUnits(wire[:0], prices[0])
```

Use `FixedDecimal` where retained canonical text and typed value methods are useful;
use `[]uint64` or `[]uint256.Int` for large numeric-only working sets. Both
forms use the same strict parser and canonical formatter.

## Decimal places and storage range

Fractional decimal places and integer capacity are independent. A
one-decimal-place price can be `25.5` or `1844674407370955161.5`; decimal
places alone cannot select a safe backend. Choose an explicit backend when a
narrower range is part of the contract:

```go
type SmallPriceFormat = sailfish.PriceInUint16Units[sailfish.DecimalPlaces2]
type SmallPrice = sailfish.FixedDecimal[SmallPriceFormat, uint16]

codec, err := sailfish.NewFixedDecimalCodec[SmallPriceFormat]()
if err != nil {
	return err
}
price, err := codec.Parse("655.35")
// codec.MaxIntegerDigits() == 3
```

The generic format families are:

```text
PriceInUint{8,16,32,64,256}Units[DecimalPlacesN]
AmountInUint{8,16,32,64,256}Units[DecimalPlacesN]
```

Price and amount formats remain different types even when backend and decimal
places match. `DecimalPlaces0` through `DecimalPlaces20` are provided; custom
zero-sized types can represent other supported decimal-place counts.

| Backend | Maximum units | Maximum fractional decimal places | Maximum decimal digits |
|---|---:|---:|---:|
| `uint8` | `255` | 2 | 3 |
| `uint16` | `65535` | 4 | 5 |
| `uint32` | `4294967295` | 9 | 10 |
| `uint64` | `18446744073709551615` | 19 | 20 |
| `uint256.Int` | `2^256 - 1` | 77 | 78 |

`FixedDecimalCodec.MaxIntegerDigits` reports the maximum integer-part digit count for a
format. It is a capacity description, not a promise that every number with
that many digits fits the binary backend.

There is one format API: `PriceInUint*Units[DecimalPlacesN]` and
`AmountInUint*Units[DecimalPlacesN]`. Cached `FixedDecimalCodec` operations
resolve fractional decimal places once; use a codec on hot paths. Each generic
format embeds a concrete backend, so it does not pay generic backend dispatch.

## Wide values

The common 18-decimal on-chain amount is explicit:

```go
type AmountFormat = sailfish.AmountInUint256Units[sailfish.DecimalPlaces18]
type Amount = sailfish.FixedDecimal[AmountFormat, uint256.Int]

amountCodec, err := sailfish.NewFixedDecimalCodec[AmountFormat]()
if err != nil {
	return err
}
```

The format selects semantic kind, fractional decimal places, and unit backend. The
sealed unit-provider interface prevents pairing a format with the wrong unit
type.

When trusted venue metadata resolves fractional decimal places at runtime,
use the concrete `Uint256FixedDecimalCodec` to avoid generic format dispatch:

```go
codec, err := sailfish.NewUint256FixedDecimalCodec(6)
if err != nil {
	return err
}

var units uint256.Int
if err := codec.ParseInto("123.456789", &units); err != "" {
	return err
}

dst := codec.AppendTo(make([]byte, 0, 32), units)
```

`Uint256FixedDecimalCodec` stores the validated fractional decimal places once.
It does not attach metadata to each value; callers remain responsible for
selecting the codec from canonical venue metadata.

## Parsing and ownership

Parsing is strict. Constructors do not trim input and do not accept signs,
exponents, missing integer/fraction digits, or excess precision.

```go
codec.Parse(s)        // retains s only when it is already canonical
codec.ParseCompact(s) // never retains s
codec.ParseBytes(b)   // parses bytes directly and never retains them
```

Non-canonical accepted input is normalized only while constructing the value:

```text
"001.2" with 5 fractional decimal places -> "1.20000"
```

Use `ParseCompact` when a short input string may reference a much larger
response buffer.

## Immutable string representation

`FixedDecimal` may retain canonical wire text:

```go
representation string
```

A string header is smaller than a byte-slice header and `String` can return it
without conversion or allocation. A Go string cannot be safely extended or
edited in place. Arithmetic therefore updates units and invalidates only the
header:

```go
d.units = newUnits
d.representation = ""
```

This invalidation allocates nothing. Any string returned before the mutation
remains immutable and valid. After mutation:

- `AppendTo` remains allocation-free when its destination has capacity.
- `String` creates one newly owned string.
- `Canonical` returns a copy that retains that string once.

There is no mutable lazy cache, so concurrent reads do not race.

## Core API

| Need | API |
|---|---|
| Parse retained canonical text | `NewFixedDecimal`, `FixedDecimalCodec.Parse` |
| Parse without retaining input | `NewCompactFixedDecimal`, `NewFixedDecimalFromBytes`, codec equivalents |
| Construct/read scaled units | `NewFixedDecimalFromUnits`, `FixedDecimalCodec.FromUnits`, `Units` |
| Parse/format compact unit batches | `FixedDecimalCodec.ParseUnits`, `FixedDecimalCodec.ParseUnitsBytes`, `FixedDecimalCodec.AppendUnits`, `FixedDecimalCodec.UnitsLen` |
| Validate and cache a static format | `NewFixedDecimalCodec`, `FixedDecimalCodec.FractionalDecimalPlaces`, `FixedDecimalCodec.MaxIntegerDigits` |
| Replace or inspect value state | `SetUnits`, `IsZero`, `HasRepresentation` |
| Exact encoded lengths | `Len`, `CBORLen`, codec equivalents |
| Runtime-scale uint256 text | `NewUint256FixedDecimalCodec`, `Parse`, `ParseBytes`, `ParseInto`, `ParseBytesInto`, `AppendTo` |
| Caller-buffer serialization | `AppendTo`, `AppendJSON`, `AppendText` |
| Caller-buffer CBOR | `AppendCBOR`, `FixedDecimalCodec.AppendCBOR`, `Uint256FixedDecimalCodec.AppendCBOR` |
| Strict CBOR decode | `UnmarshalCBOR`, `FixedDecimalCodec.ParseCBOR`, `Uint256FixedDecimalCodec.ParseCBOR`, `Uint256FixedDecimalCodec.ParseCBORInto` |
| Positional-array CBOR decode | `FixedDecimalCodec.ParseCBORFirst`, `Uint256FixedDecimalCodec.ParseCBORFirst`, `Uint256FixedDecimalCodec.ParseCBORFirstInto` |
| Owned or retained serialization | `String`, `Canonical`, `MarshalText`, `MarshalJSON`, `MarshalCBOR` |
| Same-format ordering | `Compare`, `Cmp`, `Equal`, `Less` methods |
| Cross-scale/backend ordering | package-level `Compare` |
| Checked arithmetic | `Add`, `Sub`, `AddAssign`, `SubAssign` |
| Overflow-style arithmetic | `AddOverflow`, `SubUnderflow` |

Use `FixedDecimalCodec[V, U]` for repeated work with a compile-time format. Its zero value
is valid; `NewFixedDecimalCodec` additionally validates and caches the format metadata.
Use `Uint256FixedDecimalCodec` when trusted metadata supplies fractional
decimal places at runtime.
Its `Into` methods leave the destination unchanged on error. Invalid formats,
inputs, and arithmetic return errors or status values; the package does not
use panics as an API contract.

## Serialization and deserialization

Sailfish exposes owned standard interfaces and caller-buffer APIs. Use the
owned forms at ordinary application boundaries and append/prefix-decode forms
for MDBX records, network frames, and other hot aggregate codecs.

| Format | Encode | Decode | Wire contract |
|---|---|---|---|
| Canonical text | `AppendText`, `MarshalText` | `UnmarshalText` | Exact fixed-scale ASCII decimal |
| JSON | `AppendJSON`, `MarshalJSON` | `UnmarshalJSON` | Quoted decimal string; bare numbers rejected |
| CBOR scalar | `AppendCBOR`, `MarshalCBOR` | `ParseCBOR`, `UnmarshalCBOR` | Preferred unsigned integer or tag-2 bignum |
| Positional CBOR | repeated `AppendCBOR` | repeated `ParseCBORFirst` | FixedDecimal scalars inside a parent `toarray` record |

### Text and JSON

JSON values are quoted decimal strings. Bare JSON numbers are rejected. JSON
integration and escaped-string decoding use
[`github.com/goccy/go-json`](https://github.com/goccy/go-json); ordinary
unescaped decimal strings decode directly from the JSON input.

```go
text, err := price.MarshalText() // []byte("123.31232")
if err != nil {
	return err
}
jsonValue, err := price.MarshalJSON() // []byte("\"123.31232\"")
if err != nil {
	return err
}

var decoded Price
if err := decoded.UnmarshalJSON(jsonValue); err != nil {
	return err
}
```

`AppendText` and `AppendJSON` reuse caller capacity. `MarshalText` and
`MarshalJSON` return owned slices and therefore allocate their result once.
The direct JSON decoder parses ordinary quoted decimals from the input bytes
without allocating. Escaped JSON strings take the standards-compliant
`go-json` unescape path before decimal parsing.

For a hot aggregate encoder, reuse caller capacity rather than asking each
field for an owned result:

```go
wire := make([]byte, 0, 32)
wire = price.AppendJSON(wire[:0])
```

`MarshalJSON` reserves the exact native/retained size. For an unretained
`uint256.Int`, it reserves the bounded maximum and performs the expensive
wide decimal split only once.

### Compact deterministic CBOR

CBOR stores only the scaled unsigned integer. The decimal scale and semantic
kind are compile-time format identity, while retained source text is cache
state; none of them is duplicated in storage.

```text
FixedDecimal[PriceInUint64Units[DecimalPlaces5]] units 12331232 -> 1a00bc28e0
FixedDecimal[AmountInUint256Units[DecimalPlaces18]] units 2^64 -> c249010000000000000000
```

Native values use RFC 8949's shortest unsigned-integer representation. A
`uint256.Int` uses the same representation while it fits in `uint64`, then tag
2 with a minimal big-endian magnitude. Decode accepts only preferred,
definite-length encodings and rejects trailing data, longer integer forms,
leading-zero bignums, and values outside the selected unit backend.

Sailfish decimals implement `MarshalCBOR` and `UnmarshalCBOR` for
`github.com/fxamacker/cbor/v2`. They remain scalar elements inside a compact
parent array:

```go
type Quote struct {
	_ struct{} `cbor:",toarray"`

	Price  sailfish.FixedDecimal[PriceFormat, uint64]
	Amount sailfish.FixedDecimal[AmountFormat, uint256.Int]
}
```

This encodes as `[priceUnits, amountUnits]`, not nested one-element arrays.
Wire sizes are 1-9 bytes for native units and 1-35 bytes for `uint256` units,
before the enclosing array header.

Cache fxamacker modes for reflective or cold-path aggregate encoding:

```go
enc, err := cbor.CanonicalEncOptions().EncMode()
if err != nil {
	return err
}
dec, err := cbor.DecOptions{}.DecMode()
if err != nil {
	return err
}

raw, err := enc.Marshal(quote)
if err != nil {
	return err
}
var decoded Quote
if err := dec.Unmarshal(raw, &decoded); err != nil {
	return err
}
```

Use `AppendCBOR` or the cached codec equivalent when building a hot MDBX value:

```go
dst := make([]byte, 0, 1+2*sailfish.MaxCBORSize)
dst = append(dst, 0x82) // fixed two-field CBOR array
dst = priceCodec.AppendCBOR(dst, price)
dst = amountCodec.AppendCBOR(dst, amount)
```

Decode decimal fields from a manual positional array without first finding or
copying each scalar item:

```go
price, raw, err := priceCodec.ParseCBORFirst(raw)
if err != nil {
	return err
}
amount, raw, err := amountCodec.ParseCBORFirst(raw)
if err != nil {
	return err
}
```

`ParseCBORFirst` consumes exactly one preferred deterministic unsigned value
and returns the unconsumed suffix. `ParseCBOR` remains the whole-item API and
rejects trailing data. On failure, prefix decoders return no suffix and
`ParseCBORFirstInto` leaves its destination unchanged.

The hot positional path must validate the parent array header and field count
at the enclosing-record layer. Sailfish then validates each scalar's preferred
encoding, backend range, and complete consumption. There is one current CBOR
format: no compatibility decoder, alternate integer form, or legacy fallback.

These append APIs and all direct decode APIs are `0 B/op, 0 allocs/op` with a
sized caller buffer. `MarshalCBOR` necessarily allocates one owned result
slice. The reflective `fxamacker` parent marshal also invokes that owned-slice
interface for each decimal; use the append path when aggregate encoding must
remain allocation-free. Cache fxamacker `EncMode` and `DecMode` for generic or
cold paths; they are configured codec instances, not interfaces implemented by
application values. Reflective `toarray` decode remains allocation-free for the
tested fixed quote shape.

A permanent fourteen-field oracle test builds a 93-byte positional record with
cached Sailfish codecs and verifies byte-for-byte equality with deterministic
fxamacker `cbor:",toarray"`. The 93-byte result belongs to that synthetic value
set; it is neither a fixed record size nor a theoretical minimum. The same
schema has a 15-byte structural floor when its symbol is empty and every
numeric value is zero. It has no finite format-wide maximum until the enclosing
record bounds symbol length.

A separate July 15, 2026 snapshot covers 100 MEXC spot, 100 Hyperliquid spot,
and 100 Hyperliquid perpetual markets, ranked by reported 24-hour volume. It
uses distinct positive price and quantity observations from venue metadata,
context, ticker, and L2 book responses. Each market identity appears once in
the fixture; observed values are deduplicated before min/quantile/max selection.
For realistic nonzero records, the resulting 14-field wires are 48-78 bytes:

| Cohort | Quantity case | Min | p50 | p95 | Max | Mean |
|---|---|---:|---:|---:|---:|---:|
| MEXC spot | min / median / max | 55 / 59 / 62 | 60 / 63 / 64 | 68 / 71 / 74 | 75 / 78 / 78 | 60.96 / 63.66 / 65.92 |
| Hyperliquid spot | min / median / max | 48 / 48 / 48 | 56 / 60 / 60 | 64 / 66 / 68 | 69 / 71 / 73 | 57.42 / 60.18 / 60.80 |
| Hyperliquid perps | min / median / max | 53 / 56 / 57 | 56 / 60 / 60 | 64 / 67 / 68 | 66 / 68 / 69 | 57.86 / 60.61 / 61.52 |

The snapshot and its invariant checks live in
`testdata/market_cbor_samples.json` and `market_cbor_benchmark_test.go`. Its
direct encode path is allocation-free; decode allocates only when the parent
record must own a symbol string. Sailfish numeric field decode remains
allocation-free. See [BENCHMARKS.md](BENCHMARKS.md#real-market-cbor-records)
for source policy and repeated timing results.

## Errors

Errors are typed string constants:

```go
type Error string

const ErrSyntax Error = "sailfish: invalid syntax"
```

They are comparable, allocation-free to return, and work with `errors.Is`.
Sailfish does not expose panic-on-error constructors. Unsupported fractional
decimal places and invalid input are returned as errors. A zero
`FixedDecimalCodec[V, U]` derives valid compile-time decimal places; a zero
`Uint256FixedDecimalCodec` represents zero fractional decimal places.

## Range model

The complete digit sequence is one scaled integer, so fractional decimal
places consume integer range.

| Backend | Maximum fractional decimal places | Raw units |
|---|---:|---:|
| `uint8` | 2 | 1 byte |
| `uint16` | 4 | 2 bytes |
| `uint32` | 9 | 4 bytes |
| `uint64` | 19 | 8 bytes |
| `uint256.Int` | 77 | 32 bytes |

On 64-bit systems:

```text
FixedDecimal[..., uint8]        24 bytes
FixedDecimal[..., uint16]       24 bytes
FixedDecimal[..., uint32]       24 bytes
FixedDecimal[..., uint64]       24 bytes
FixedDecimal[..., uint256.Int]  48 bytes
FixedDecimalCodec                1 byte
```

Narrow native units enforce smaller ranges and reduce standalone/raw unit
arrays. They do not reduce the current `FixedDecimal` struct below 24 bytes because
its retained immutable string header and alignment dominate the layout.
The incomparability marker is the first zero-sized field so it does not create
trailing zero-field padding. Layout tests lock unit and string offsets, struct
alignment, and these sizes on 64-bit targets.

## Deliberate boundaries

The initial package does not define:

- signed decimals;
- implicit truncation or rounding;
- multiplication or division rounding policy;
- floating-point conversion;
- mutable shared caches;
- a runtime-varying scale carried by every value.

Those are separate financial contracts, not parser conveniences. Custom
zero-sized `StaticDecimalPlaces` types cover compile-time scales beyond `DecimalPlaces20`.

## Performance

These are five-run summaries from a complete `make bench` execution on Go
1.26.5, darwin/arm64, Apple M1 Max. Microbenchmark numbers are local, not
portable guarantees; compare changes on the same host and toolchain.

### Parsing and formatting

| Operation | Time | B/op | allocs/op |
|---|---:|---:|---:|
| Parse canonical `uint64` through `FixedDecimalCodec` | 7.75 ns | 0 | 0 |
| Parse canonical `uint256.Int` | 49.2 ns | 0 | 0 |
| Parse maximum 78-digit `uint256.Int` | 64.6 ns | 0 | 0 |
| Append retained `uint64` | 2.90 ns | 0 | 0 |
| Append formatted `uint64` | 9.8 ns | 0 | 0 |
| Append formatted four-limb `uint256.Int` | 112 ns | 0 | 0 |
| Return retained `String` | 2.12 ns | 0 | 0 |
| Return newly formatted `String` | 27.1 ns | 16 | 1 |

### Width scaling

| Dense parse kernel | 19 digits | 38 digits | 57 digits | 77 digits |
|---|---:|---:|---:|---:|
| `string` input | 9.56 ns | 18.9 ns | 28.5 ns | 43.3 ns |
| `[]byte` input | 9.39 ns | 18.2 ns | 28.0 ns | 42.8 ns |

| Formatted width | One limb | Two limbs | Three limbs | Four limbs | Maximum |
|---|---:|---:|---:|---:|---:|
| Wide formatting kernel | 17.2 ns | 46.0 ns | 72.9 ns | 107 ns | 131 ns |

Every width-scaling parse and append row above is `0 B/op`, `0 allocs/op`.

### Comparison and arithmetic

| Operation | Time | B/op | allocs/op |
|---|---:|---:|---:|
| Same-scale `uint64` compare | 2.10 ns | 0 | 0 |
| Same-scale `uint256.Int` compare | 6.37 ns | 0 | 0 |
| Cross-scale/backend compare | 50.7 ns | 0 | 0 |
| Checked `uint64` add-assign | 4.38 ns | 0 | 0 |
| Checked `uint256.Int` add-assign | 13.2 ns | 0 | 0 |

### Serialization

| Operation | Time | B/op | allocs/op |
|---|---:|---:|---:|
| Append retained / formatted native JSON | 4.41 / 14.4 ns | 0 | 0 |
| Append retained / formatted wide JSON | 7.31 / 141 ns | 0 | 0 |
| Owned native retained / formatted `MarshalJSON` | 20.1 / 34.9 ns | 16 | 1 |
| Owned wide retained / formatted `MarshalJSON` | 32.4 / 181 ns | 96 | 1 |
| Unmarshal canonical native / wide JSON | 14.4 / 78.8 ns | 0 | 0 |
| Unmarshal escaped native JSON | 120 ns | 40 | 2 |
| Append native / `uint256` CBOR scalar | 3.54 / 8.03 ns | 0 | 0 |
| Decode native / `uint256` CBOR scalar | 8.07 / 12.8 ns | 0 | 0 |
| Runtime-codec `uint256` append, one limb / maximum | 4.03 / 6.52 ns | 0 | 0 |
| Runtime-codec `uint256` decode, one limb / maximum | 4.17 / 5.81 ns | 0 | 0 |
| Owned native / `uint256` `MarshalCBOR` | 20.2 / 28.3 ns | 16 / 32 | 1 |
| fxamacker two-field `toarray` marshal / unmarshal | 175 / 146 ns | 120 / 0 | 4 / 0 |
| Manual 14-field positional CBOR encode / decode | 50.1 / 93.8 ns | 0 / 8 | 0 / 1 |

The manual record decoder's one allocation owns its parent string field;
Sailfish numeric field decoding is allocation-free. Owned `String` and marshal
results allocate by contract. Detailed commands, profiles, and allocation
ownership are in [BENCHMARKS.md](BENCHMARKS.md).

For values parsed from canonical venue text, retain the representation with
`FixedDecimalCodec.Parse`; subsequent appends are below 5 ns even for a four-limb value.
Use raw-unit formatting for constructed or mutated values, and call
`Canonical` once when the same formatted value will be emitted repeatedly.

The amd64/arm64 SWAR loader uses one narrowly scoped read-only `unsafe` load;
other architectures use the byte-shift loader. The pointer is never retained
or used for mutation, and release validation includes cross-builds, race, and
`checkptr=2`. No assembly or runtime CPU-feature dispatch is used: measured
short-token latency does not justify their call and maintenance cost.

## Algorithms and measured choices

| Area | Current algorithm | Reason |
|---|---|---|
| Scale model | Zero-sized compile-time format or one-byte `Uint256FixedDecimalCodec` | Static strategies carry no runtime scale; dynamic venue metadata validates scale once |
| Numeric model | One unsigned scaled integer | Exact comparison/arithmetic with no floating-point state |
| Native parsing | Pairwise accumulation plus known-point SWAR for exact 8/16-digit shapes | Keeps irregular inputs simple while bringing the common `123.31232` parse to 7.75 ns |
| Wide parsing | One or two independent eight-digit SWAR blocks plus a scalar tail | Reduced 19-78 digit parse time by roughly 9-19% in the latest round |
| SWAR loads | One read-only unaligned native load on amd64/arm64; byte shifts elsewhere | Removes load assembly on release architectures without retaining or mutating input memory |
| Native formatting | Pairwise digits plus selected reverse-SWAR widths and branchless digit count | Arithmetic-packed 5-8 and 14-20 digit scaled values improve 5-29%; other widths retain the smaller pair-table path |
| Wide formatting | Base-`1e19` chunks using precomputed-reciprocal 2-by-1 division | Avoids serial hardware division and reduced two-to-four-limb formatting by roughly 9-24% |
| Repeated output | Retain immutable canonical input or call `Canonical` once | Repeated append becomes a short string copy |
| JSON | Direct quoted append and parse-first unescaped decode | Keeps canonical JSON encode/decode allocation-free; escaped input uses the standards-compliant slow path |
| CBOR | Preferred unsigned integer; tag 2 only above `uint64` | Small deterministic wire with strict decoding |
| Hot aggregate CBOR | Caller-buffer scalar append and positional prefix decode | Avoids reflection and owned per-field slices |
| Type dispatch | Concrete backend embedded in each format; cached scale in `FixedDecimalCodec` | Avoids generic backend type switches and repeated metadata work |
| Errors | Pre-boxed typed string constants | Comparable errors with zero per-call failure allocation |
| Numeric batch locality | Contiguous raw units with `ParseUnits` / `AppendUnits` | Avoids scanning optional text-cache state when only numeric units are used |
| Profile-guided optimization | Production CPU profile owned by the consuming `main` package | PGO optimizes the whole executable; Sailfish does not ship a benchmark-trained library profile |

Measured alternatives are not retained in production: base-`1e9` and
base-`1e16` wide formatting were slower outside narrow synthetic cases, direct
decimal placement across wide chunks was 2-5% slower, applying base-`1e8`
native formatting to every width regressed protected paths, generated
per-scale masks duplicated code, overstore violated caller-buffer ownership,
and a `0-99` cache penalized representative misses. See
[PERFORMANCE.md](PERFORMANCE.md) for the benchmark artifacts and acceptance
decisions.

Cache-locality crossover benchmarks on an Apple M1 Max found no meaningful
representation difference for sequential scans, but random numeric scans above
the measured L2 were 1.6x faster with `[]uint64` than native `FixedDecimal` values
and 3.7x faster with `[]uint256.Int` than wide `FixedDecimal` values. These are
working-set measurements, not claimed hardware miss rates; the local CLI CPU
counter tools exposed no configured cache-miss event. The package therefore
keeps the existing minimum `FixedDecimal` layout and exposes raw-unit boundary
methods instead of adding padding, indirection, or another decimal type.

### Profile-guided application builds

Sailfish does not ship `default.pgo`. Go PGO is a whole-program optimization,
so a profile captured from this library's benchmarks cannot represent a
consumer's market mix, request frequencies, dependencies, or error paths.

Collect CPU profiles from the deployed executable, merge equal-duration
samples when needed, then compare the same revision with and without PGO:

```sh
go tool pprof -proto profile-a.pprof profile-b.pprof > default.pgo
go build -pgo=off -o app-no-pgo ./cmd/app
go build -pgo=default.pgo -o app-pgo ./cmd/app
```

A local mixed Sailfish experiment found 5-23% gains on several parse and wide
decode paths, but 8-9% regressions on native CBOR and JSON decode. That profile
is deliberately not part of the package. See [PERFORMANCE.md](PERFORMANCE.md)
for the matrix and compiler diagnostics. Consumers should gate PGO on their
whole-application CPU/latency, protected minority paths, binary size, build
time, and normal correctness suites.

### Branch prediction and assembly

Sailfish uses branchless arithmetic selectively, after distribution-level and
public-operation measurements. Native decimal width is derived from
`bits.Len64`, a fixed-point binary-to-decimal estimate, and one borrow-bit
threshold correction. On arm64, the compiler emits `CLZ`, `MUL`, `LSR`, and
`SUBS`/`NGC` instead of the former data-dependent comparison tree.

The isolated digit-width kernel improved by 7-29% across mixed-width,
fixed-width, and market-shaped inputs. Thirty paired, order-alternated runs of
the public formatted-`uint64` append path improved from 12.9 ns to 12.6 ns
(-1.98%, p=0.000), with 0 B/op and 0 allocs/op. The complete native-format
width matrix remained neutral outside that aggregate public-path gain.

The next formatter round adopted an Abseil-style reverse-SWAR conversion only
for measured 5-8 and 14-20 digit scaled widths. A rotated width bitset avoids
the guards emitted for variable shifts, packed point insertion avoids a
temporary copy for eight-digit values, and explicit BCE proofs collapse two
range checks into one. The common `123.31232` append improved from 12.6 ns to
9.8 ns (-22.2%); selected widths improved 5-29%, with 0 B/op and 0 allocs/op.
Adjacent pair-table widths remain within roughly 0-1.3% of the prior path.

Architecture-specific AVX-512 IFMA/VBMI formatting remains out of scope: the
[published algorithm](https://arxiv.org/abs/2604.26019) cannot be executed on
the arm64 release host, while the selected scalar kernels already complete in
about 7-14 ns. The portable comparison also includes Go's
[pair-table formatter](https://go.dev/src/strconv/itoa.go), retained for widths
where its smaller dependency chain wins.

Two broader branchless changes were rejected. A four-threshold arithmetic
CBOR-length calculation was 13-15% slower than the predictable switch, and
`segmentio/asm/ascii.ValidString` validates ASCII rather than Sailfish decimal
grammar, adds a second pass, and uses its generic Go path on darwin/arm64.
Existing inlined SWAR digit validation remains the better short-token kernel.

Hardware branch-miss rates are not claimed: this host has no Linux `perf`, and
the available Docker VM exposes no CPU PMU. The assembly and timing evidence,
rejected candidates, and a PMU-enabled Linux follow-up command are recorded in
[PERFORMANCE.md](PERFORMANCE.md).

## Validation

```sh
make test
make vet
make race
make bench
make fuzz
```

Tests include exhaustive byte validation, maximum-value boundaries, randomized
exact-reference properties, ownership/cache behavior, allocation assertions,
external-package API checks, and fuzz targets for both unit backends and JSON.

If this repository is cloned under a parent directory containing an unrelated
`go.work`, use `GOWORK=off` or the included Makefile.

## License

MIT. See [LICENSE](LICENSE).
