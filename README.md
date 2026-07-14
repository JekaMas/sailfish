# sailfish

`sailfish` is a fast, unsigned, fixed-scale decimal package for trading and
financial protocols that exchange exact values as strings such as
`"123.31232"`.

The numeric state is one scaled integer:

```text
value = units / 10^venue scale
```

Supported unit backends are `uint8`, `uint16`, `uint32`, `uint64`, and
`uint256.Int`. The common parse, append, compare, and arithmetic paths perform
no heap allocations.

Sailfish requires Go 1.26.5 or newer.

## Quick start

Select semantic kind, unit capacity, and fractional scale explicitly:

```go
codec, err := sailfish.NewCodec[sailfish.PriceUint64[sailfish.Fraction5]]()
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

For `PriceUint64[Fraction9]`, the maximum representable value is
`18446744073.709551615`.

## Scale and storage range

Fractional scale and integer capacity are independent. A scale-1 price can be
`25.5` or `1844674407370955161.5`; scale alone cannot select a safe backend.
Choose an explicit backend when a narrower range is part of the contract:

```go
type SmallPriceFormat = sailfish.PriceUint16[sailfish.Fraction2]
type SmallPrice = sailfish.Decimal[SmallPriceFormat, uint16]

codec, err := sailfish.NewCodec[SmallPriceFormat]()
if err != nil {
	return err
}
price, err := codec.Parse("655.35")
// codec.MaxIntegerDigits() == 3
```

The generic format families are:

```text
PriceUint8/16/32/64/256[FractionN]
AmountUint8/16/32/64/256[FractionN]
```

Price and amount formats remain different types even when backend and scale
match. `Fraction0` through `Fraction20` are provided; custom zero-sized scale
types can represent other supported scales.

| Backend | Maximum units | Maximum scale | Maximum decimal digits |
|---|---:|---:|---:|
| `uint8` | `255` | 2 | 3 |
| `uint16` | `65535` | 4 | 5 |
| `uint32` | `4294967295` | 9 | 10 |
| `uint64` | `18446744073709551615` | 19 | 20 |
| `uint256.Int` | `2^256 - 1` | 77 | 78 |

`Codec.MaxIntegerDigits` reports the maximum integer-part digit count for a
format. It is a capacity description, not a promise that every number with
that many digits fits the binary backend.

There is one format API: `PriceUint*` and `AmountUint*`. Cached `Codec`
operations resolve scale once and benchmark equivalently to a test-local
concrete venue; use a codec on hot paths. Each generic format embeds a concrete
backend, so it does not pay generic backend dispatch.

## Wide values

The common 18-decimal on-chain amount is explicit:

```go
type AmountFormat = sailfish.AmountUint256[sailfish.Fraction18]
type Amount = sailfish.Decimal[AmountFormat, uint256.Int]

amountCodec, err := sailfish.NewCodec[AmountFormat]()
if err != nil {
	return err
}
```

The format selects semantic kind, fractional scale, and unit backend. The
sealed unit-provider interface prevents pairing a format with the wrong unit
type.

When trusted venue metadata resolves a scale at runtime, use the concrete
`Uint256Codec` to avoid generic venue dispatch:

```go
codec, err := sailfish.NewUint256Codec(6)
if err != nil {
	return err
}

var units uint256.Int
if err := codec.ParseInto("123.456789", &units); err != "" {
	return err
}

dst := codec.AppendTo(make([]byte, 0, 32), units)
```

`Uint256Codec` stores the validated scale once. It does not attach a dynamic
scale to each value; callers remain responsible for selecting the codec from
canonical venue metadata.

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
"001.2" at scale 5 -> "1.20000"
```

Use `ParseCompact` when a short input string may reference a much larger
response buffer.

## Immutable string representation

`Decimal` may retain canonical wire text:

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
| Parse retained canonical text | `New`, `Codec.Parse` |
| Parse without retaining input | `NewCompact`, `NewBytes`, codec equivalents |
| Construct/read scaled units | `NewFromUnits`, `Codec.FromUnits`, `Units` |
| Inspect backend integer capacity | `Codec.MaxIntegerDigits` |
| Runtime-scale uint256 boundary | `Uint256Codec.Parse`, `ParseInto`, `AppendTo` |
| Caller-buffer serialization | `AppendTo`, `AppendJSON`, `AppendText` |
| Caller-buffer CBOR | `AppendCBOR`, `Codec.AppendCBOR`, `Uint256Codec.AppendCBOR` |
| Strict CBOR decode | `UnmarshalCBOR`, `Codec.ParseCBOR`, `Uint256Codec.ParseCBOR` |
| Positional-array CBOR decode | `Codec.ParseCBORFirst`, `Uint256Codec.ParseCBORFirst`, `ParseCBORFirstInto` |
| Owned serialization | `String`, `MarshalText`, `MarshalJSON` |
| Same-venue ordering | `Compare`, `Cmp`, `Equal`, `Less` methods |
| Cross-scale/backend ordering | package-level `Compare` |
| Checked arithmetic | `Add`, `Sub`, `AddAssign`, `SubAssign` |
| Overflow-style arithmetic | `AddOverflow`, `SubUnderflow` |

JSON values are quoted decimal strings. Bare JSON numbers are rejected.
JSON integration and escaped-string decoding use
[`github.com/goccy/go-json`](https://github.com/goccy/go-json); Sailfish's
ordinary unescaped decimal-string decode path remains allocation-free.

## Compact deterministic CBOR

CBOR stores only the scaled unsigned integer. The decimal scale and semantic
kind are compile-time format identity, while retained source text is cache
state; none of them is duplicated in storage.

```text
Decimal[PriceUint64[Fraction5]] units 12331232 -> 1a00bc28e0
Decimal[AmountUint256[Fraction18]] units 2^64 -> c249010000000000000000
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

	Price  sailfish.Decimal[PriceFormat, uint64]
	Amount sailfish.Decimal[AmountFormat, uint256.Int]
}
```

This encodes as `[priceUnits, amountUnits]`, not nested one-element arrays.
Wire sizes are 1-9 bytes for native units and 1-35 bytes for `uint256` units,
before the enclosing array header.

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
fxamacker `cbor:",toarray"`. Its direct encode path is allocation-free; decode
allocates only the owned string field in the parent record. Sailfish numeric
field decode remains allocation-free.

## Errors

Errors are typed string constants:

```go
type Error string

const ErrSyntax Error = "sailfish: invalid syntax"
```

They are comparable, allocation-free to return, and work with `errors.Is`.
Sailfish does not expose panic-on-error constructors. Invalid scale and input
configuration are returned as errors. A zero `Codec[V, U]` derives the valid
compile-time venue scale; a zero `Uint256Codec` is the useful scale-0 codec.

## Range model

The complete digit sequence is one scaled integer, so scale consumes integer
range.

| Backend | Maximum scale | Raw units |
|---|---:|---:|
| `uint8` | 2 | 1 byte |
| `uint16` | 4 | 2 bytes |
| `uint32` | 9 | 4 bytes |
| `uint64` | 19 | 8 bytes |
| `uint256.Int` | 77 | 32 bytes |

On 64-bit systems:

```text
Decimal[..., uint8]        24 bytes
Decimal[..., uint16]       24 bytes
Decimal[..., uint32]       24 bytes
Decimal[..., uint64]       24 bytes
Decimal[..., uint256.Int]  48 bytes
Codec                       1 byte
```

Narrow native units enforce smaller ranges and reduce standalone/raw unit
arrays. They do not reduce the current `Decimal` struct below 24 bytes because
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
zero-sized `VenueScale` types cover compile-time scales beyond `Fraction20`.

## Performance

Representative local results on Go 1.26.5, darwin/arm64, Apple M1 Max:

| Operation | Approximate time | B/op | allocs/op |
|---|---:|---:|---:|
| Parse canonical scale-5 `uint64` | 9.79 ns | 0 | 0 |
| Parse runtime scale-6 `uint256.Int` into caller storage | 10.2 ns | 0 | 0 |
| Parse canonical 38-digit `uint256.Int` | 51.7 ns | 0 | 0 |
| Append retained `uint64` | 2.73 ns | 0 | 0 |
| Append retained four-limb `uint256.Int` | 4.22 ns | 0 | 0 |
| Append formatted `uint64` | 13.7 ns | 0 | 0 |
| Append formatted one-limb `uint256.Int` | 13.1 ns | 0 | 0 |
| Append formatted four-limb `uint256.Int` | 138-152 ns | 0 | 0 |
| Same-scale `uint64` compare | 2.14 ns | 0 | 0 |
| Cross-scale/backend compare | 52.7 ns | 0 | 0 |
| Formatted owned `String` | 32.8 ns | 16 | 1 |

The one `String` allocation is the returned string's ownership contract.
Detailed commands and profile interpretation are in [BENCHMARKS.md](BENCHMARKS.md).

For values parsed from canonical venue text, retain the representation with
`Codec.Parse`; subsequent appends are below 5 ns even for a four-limb value.
Use raw-unit formatting for constructed or mutated values, and call
`Canonical` once when the same formatted value will be emitted repeatedly.

No `unsafe` or assembly is used in production code. Profiles show the pure-Go
implementation is already allocation-free on hot paths; adding
architecture-specific code at this point would add maintenance and audit risk
without a demonstrated target.

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
