# sailfish

`sailfish` is a fast, unsigned, fixed-scale decimal package for trading and
financial protocols that exchange exact values as strings such as
`"123.31232"`.

The numeric state is one scaled integer:

```text
value = units / 10^venue scale
```

Supported unit backends are `uint64` and `uint256.Int`. The common parse,
append, compare, and arithmetic paths perform no heap allocations.

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

`PriceScale1` through `PriceScale9` use `uint64`. At scale 9, the maximum
representable value is `18446744073.709551615`.

## Wide values

Define a zero-sized venue that embeds `Uint256Units`:

```go
type AmountScale18 struct{ sailfish.Uint256Units }

func (AmountScale18) NotionScale() sailfish.Notion { return 18 }

type Amount = sailfish.Decimal[AmountScale18, uint256.Int]

var amountCodec = sailfish.MustCodec[AmountScale18]()
```

The venue selects both scale and unit backend. The sealed unit-provider
interface prevents accidentally pairing a venue with the wrong unit type.

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

| Backend | Maximum scale | Storage |
|---|---:|---:|
| `uint64` | 19 | 8 bytes |
| `uint256.Int` | 77 | 32 bytes |

On 64-bit systems:

```text
Decimal[..., uint64]       24 bytes
Decimal[..., uint256.Int]  48 bytes
Codec                       1 byte
```

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

Representative local results on Go 1.26.2, darwin/arm64, Apple M1 Max:

| Operation | Approximate time | B/op | allocs/op |
|---|---:|---:|---:|
| Parse canonical scale-5 `uint64` | 11.7 ns | 0 | 0 |
| Parse canonical scale-18 `uint256.Int` | 52.6 ns | 0 | 0 |
| Append retained `uint64` | 2.9 ns | 0 | 0 |
| Append formatted `uint64` | 13.5 ns | 0 | 0 |
| Append formatted wide `uint256.Int` | 148 ns | 0 | 0 |
| Same-scale `uint64` compare | 2.1 ns | 0 | 0 |
| Cross-scale/backend compare | 52 ns | 0 | 0 |
| Formatted owned `String` | 30.1 ns | 16 | 1 |

The one `String` allocation is the returned string's ownership contract.
Detailed commands and profile interpretation are in [BENCHMARKS.md](BENCHMARKS.md).

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
