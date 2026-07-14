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

The package provides ready-to-use uint64 price scales from 1 through 9:

```go
codec := sailfish.MustCodec[sailfish.PriceScale5]()

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

`PriceScale1` through `PriceScale9` describe only the number of fractional
digits. Their ready-to-use defaults retain the broad `uint64` range. At scale
9, the maximum representable value is `18446744073.709551615`.

## Scale and storage range

Fractional scale and integer capacity are independent. A scale-1 price can be
`25.5` or `1844674407370955161.5`; scale alone cannot select a safe backend.
Choose an explicit backend when a narrower range is part of the contract:

```go
type SmallPriceFormat = sailfish.PriceUint16[sailfish.Fraction2]
type SmallPrice = sailfish.Decimal[SmallPriceFormat, uint16]

codec := sailfish.MustCodec[SmallPriceFormat]()
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

The built-in `PriceScale1` through `PriceScale9` remain concrete `uint64`
formats. Benchmarks found generic scale metadata equivalent for cached
`Codec.Parse`, but about 1 ns slower for one-shot `New` and direct
`Decimal.AppendTo`. Explicit `PriceUint*` and `AmountUint*` formats provide
generic composition without adding a generic backend dispatch.

## Wide values

The common 18-decimal on-chain amount is built in:

```go
type Amount = sailfish.Decimal[sailfish.AmountScale18, uint256.Int]

var amountCodec = sailfish.MustCodec[sailfish.AmountScale18]()
```

Other amount scales and ranges use the generic `AmountUint*` formats. The
format selects semantic kind, fractional scale, and unit backend. The sealed
unit-provider interface prevents pairing a format with the wrong unit type.

When trusted venue metadata resolves a scale at runtime, use the concrete
`Uint256Codec` to avoid generic venue dispatch:

```go
codec := sailfish.MustUint256Codec(6)

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
| Owned serialization | `String`, `MarshalText`, `MarshalJSON` |
| Same-venue ordering | `Compare`, `Cmp`, `Equal`, `Less` methods |
| Cross-scale/backend ordering | package-level `Compare` |
| Checked arithmetic | `Add`, `Sub`, `AddAssign`, `SubAssign` |
| Overflow-style arithmetic | `AddOverflow`, `SubUnderflow` |

JSON values are quoted decimal strings. Bare JSON numbers are rejected.
JSON integration and escaped-string decoding use
[`github.com/goccy/go-json`](https://github.com/goccy/go-json); Sailfish's
ordinary unescaped decimal-string decode path remains allocation-free.

## Errors

Errors are typed string constants:

```go
type Error string

const ErrSyntax Error = "sailfish: invalid syntax"
```

They are comparable, allocation-free to return, and work with `errors.Is`.

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

## Deliberate boundaries

The initial package does not define:

- signed decimals;
- implicit truncation or rounding;
- multiplication or division rounding policy;
- floating-point conversion;
- mutable shared caches;
- a runtime-varying scale carried by every value.

Those are separate financial contracts, not parser conveniences. Custom venue
types cover compile-time scales beyond the built-in price scales.

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
