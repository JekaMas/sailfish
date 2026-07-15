# Sailfish Guide for Coding Agents

This document is the concise operational contract for using
`github.com/JekaMas/sailfish`. It is written for coding agents and reviewers
that need to choose a decimal type, preserve exact units, and avoid accidental
allocations or wire-format drift.

Current package contract:

- release: `v1.0.4`;
- Go: `1.26.5` or newer;
- numbers: unsigned, fixed scale, integer-backed;
- numeric value: `units / 10^fractionalDecimalPlaces`;
- unit backends: `uint8`, `uint16`, `uint32`, `uint64`, `uint256.Int`;
- no floating-point conversion, signed values, rounding, or implicit
  truncation;
- no legacy codecs, compatibility aliases, alternate wire formats, or
  fallback decoders;
- invalid input returns typed errors; package APIs do not use panic as a
  contract.

The complete reference and benchmark evidence remain in [README.md](README.md),
[BENCHMARKS.md](BENCHMARKS.md), and [PERFORMANCE.md](PERFORMANCE.md).

## 1. Decide Whether Sailfish Applies

Use Sailfish when all of these are true:

1. The value is non-negative.
2. Its fractional decimal places are known from a protocol contract or trusted
   metadata.
3. Its exact value can be represented as one scaled unsigned integer.
4. Text, JSON, or CBOR must round-trip without floating-point state.

Do not use Sailfish when the domain needs:

- negative values;
- implicit rounding or truncation;
- a scale that travels with every individual value;
- multiplication or division with an unspecified rounding policy;
- compatibility decoding of an older wire representation.

Define those contracts separately before selecting a numeric implementation.

## 2. Choose Static or Runtime Decimal Places

### Static decimal places

Use `FixedDecimal` plus `FixedDecimalCodec` when fractional decimal places are
known at compile time and semantic type identity is useful.

```go
type PriceFormat = sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]
type Price = sailfish.FixedDecimal[PriceFormat, uint64]

priceCodec, err := sailfish.NewFixedDecimalCodec[PriceFormat]()
if err != nil {
	return err
}
```

The zero value of `FixedDecimalCodec` derives the static format and is valid;
prefer the constructor at an ownership boundary because it validates and caches
the format once.

The format name is the contract:

```text
PriceInUint64Units[DecimalPlaces5]
│            │         └─ exactly five fractional decimal places
│            └─────────── raw state is one uint64
└──────────────────────── semantic kind is price
```

Price and amount formats are intentionally distinct even when they use the
same backend and fractional decimal places.

Available semantic formats:

```text
PriceInUint{8,16,32,64,256}Units[DecimalPlacesN]
AmountInUint{8,16,32,64,256}Units[DecimalPlacesN]
```

`DecimalPlaces0` through `DecimalPlaces20` are predefined. A custom zero-sized
type implementing `StaticDecimalPlaces` can supply another supported value.

### Runtime decimal places

Use `Uint256FixedDecimalCodec` when trusted metadata resolves fractional
decimal places at runtime, as with heterogeneous CEX markets.

```go
codec, err := sailfish.NewUint256FixedDecimalCodec(
	sailfish.DecimalPlaces(fractionalDecimalPlaces),
)
if err != nil {
	return err
}

var units uint256.Int
if parseErr := codec.ParseBytesInto(raw, &units); parseErr != "" {
	return parseErr
}
```

Validate and cache this codec with the metadata object. Do not reconstruct it
for every value, and do not store its scale in every row.

The zero value of `Uint256FixedDecimalCodec` means zero fractional decimal
places. Metadata-driven code must use the constructor instead of treating a
missing codec as a valid default.

`Uint256FixedDecimalCodec` returns raw `uint256.Int` units. It deliberately
does not attach semantic identity, token identity, market identity, or retained
text to each value. The calling domain type must own those concerns.

## 3. Choose the Unit Backend Independently

Fractional decimal places do not determine integer capacity. Select the
narrowest backend whose complete scaled-integer range covers the protocol
contract.

| Backend | Decimal digits in maximum units | Maximum fractional decimal places |
|---|---:|---:|
| `uint8` | 3 | 2 |
| `uint16` | 5 | 4 |
| `uint32` | 10 | 9 |
| `uint64` | 20 | 19 |
| `uint256.Int` | 78 | 77 |

Use `FixedDecimalCodec.MaxIntegerDigits()` to inspect capacity after choosing a
static format. It is a capacity description, not proof that every value with
that many integer digits fits.

Prefer native units for bounded protocol values and `uint256.Int` for EVM-width
amounts or runtime market metadata where one fixed backend simplifies the
boundary contract.

## 4. Parse at a Boundary, Then Keep Units

Parsing is strict:

- no leading or trailing whitespace;
- no `+` or `-` sign;
- no exponent notation;
- no missing integer digits;
- no repeated decimal point;
- no fractional digits beyond the configured count;
- no overflow of the selected backend.

Constructors may canonicalize accepted decimal syntax, for example
`"001.2"` at five fractional decimal places becomes `"1.20000"`. Business
logic must not trim, normalize, or reparse values after construction.

Choose the parse API by ownership need:

| Need | API | Input retention |
|---|---|---|
| Repeated output of canonical string input | `codec.Parse` | Retains input only when already canonical |
| Avoid retaining a large response buffer | `codec.ParseCompact` | Never retains string |
| Parse network or DB bytes directly | `codec.ParseBytes` | Never retains or converts bytes |
| Numeric-only batch | `codec.ParseUnits` / `ParseUnitsBytes` | Returns raw units only |
| Runtime-scale `uint256` boundary | `Parse`, `ParseBytes`, `ParseInto`, `ParseBytesInto` | Raw units only |

For destination APIs, failure leaves the destination unchanged.

Do not add a permissive parser fallback. If an upstream venue emits unsupported
syntax, make that protocol decision explicit and test it instead of silently
changing Sailfish grammar.

## 5. Keep Domain Identity Outside the Codec

Sailfish owns decimal representation and exact scaled-unit operations. It does
not own:

- token or asset identity;
- base/quote market identity;
- exchange identity;
- price-source identity;
- timeframe or candle policy;
- venue order tick/lot rules;
- balances, fills, accounting, or order lifecycle.

Wrap Sailfish units in domain types when those identities matter. Never pass a
naked integer across a boundary where its token, semantic kind, or scale can be
ambiguous.

## 6. Work With Values

Construct from strict text:

```go
price, err := priceCodec.Parse("123.31232")
if err != nil {
	return err
}
```

Construct from already-scaled units without a text round-trip:

```go
price := priceCodec.FromUnits(uint64(12_331_232))
```

Read units by value:

```go
units := price.Units()
```

`uint256.Int` is an inline four-limb value, so this copy does not allocate or
share mutable `big.Int` backing storage.

Same-format operations:

```go
if price.Less(other) {
	// ...
}

sum, err := price.Add(delta)
if err != nil {
	return err
}
```

Use `AddAssign` and `SubAssign` for mutation with explicit overflow/underflow
status. Failed operations leave the receiver unchanged. Successful numeric
mutation clears retained text because the text is cache state, not numeric
state.

Use package-level `sailfish.Compare(a, b)` for exact comparison across scales
or backends. It does not rescale either integer and therefore avoids rescaling
overflow.

## 7. Format Without Accidental Ownership

Use caller-buffer APIs on hot paths:

```go
buf := make([]byte, 0, priceCodec.Len(price))
buf = priceCodec.AppendTo(buf, price)
```

For numeric-only batches:

```go
buf = priceCodec.AppendUnits(buf[:0], units)
```

Use `String()` when an owned or retained string is actually required. It
returns retained canonical text without allocation; otherwise it creates one
owned string. Use `Canonical()` once when a constructed or mutated value will
be rendered repeatedly.

Never attempt to append to or mutate the retained string. Sailfish invalidates
only its immutable string header after numeric mutation.

## 8. JSON Contract

JSON is a quoted canonical decimal string:

```json
"123.31232"
```

Bare JSON numbers are rejected. This prevents an intermediate JSON number from
losing precision or changing formatting.

Use `AppendJSON` in caller-buffer aggregate encoders. Use `MarshalJSON` and
`UnmarshalJSON` for ordinary interface-based integration. Ordinary unescaped
quoted decimals decode directly; escaped strings use the standards-compliant
`goccy/go-json` path.

## 9. CBOR Contract

CBOR stores only scaled unsigned units:

- native and one-limb values use the shortest RFC 8949 unsigned integer;
- wider `uint256.Int` values use tag 2 with a minimal big-endian magnitude;
- fractional decimal places, semantic kind, and retained text are not on the
  wire;
- non-preferred encodings, leading-zero bignums, trailing data, and backend
  overflow are rejected.

Use `MarshalCBOR` / `UnmarshalCBOR` for generic `fxamacker/cbor` integration.
They make `FixedDecimal` a scalar element inside a parent `cbor:",toarray"`
record.

```go
type Quote struct {
	_ struct{} `cbor:",toarray"`

	Price  Price
	Amount Amount
}
```

Use `AppendCBOR` and `ParseCBORFirst` for a hot manual positional record:

```go
dst = append(dst, 0x82) // parent two-field array
dst = priceCodec.AppendCBOR(dst, price)
dst = amountCodec.AppendCBOR(dst, amount)

raw := dst[1:]
price, raw, err = priceCodec.ParseCBORFirst(raw)
if err != nil {
	return err
}
amount, raw, err = amountCodec.ParseCBORFirst(raw)
if err != nil {
	return err
}
if len(raw) != 0 {
	return ErrQuoteTrailingData // typed error owned by the parent schema
}
```

The parent codec owns array header, field count, ordering, and complete-record
validation. Sailfish owns each decimal scalar. Cache `fxamacker` `EncMode` and
`DecMode` for generic cold paths; application values implement
`Marshaler`/`Unmarshaler`, not mode interfaces.

There is one CBOR scalar format. Do not add version N-1 decoders or alternate
integer/text fallbacks.

## 10. Allocation Rules

Expected allocation-free hot APIs, with sufficient destination capacity:

- static and runtime parsing;
- `ParseInto` / `ParseBytesInto`;
- `ParseUnits` / `ParseUnitsBytes`;
- `AppendTo`, `AppendUnits`, `AppendJSON`, `AppendCBOR`;
- direct CBOR decode and prefix decode;
- comparisons and checked arithmetic.

Owned result APIs allocate by contract:

- a newly formatted `String()`;
- `MarshalText`, `MarshalJSON`, and `MarshalCBOR` result slices.

Do not claim an allocation regression or improvement from intuition. Add a
focused benchmark with `-benchmem` and use escape analysis or a memory profile
to identify the owner.

For large numeric-only working sets, prefer contiguous `[]uint64` or
`[]uint256.Int` and boundary codecs over arrays of `FixedDecimal`, whose
optional retained-string header increases cache footprint.

## 11. Error Handling

Exported errors are comparable typed string constants:

- `ErrSyntax`;
- `ErrRange`;
- `ErrPrecision`;
- `ErrUnsupportedFractionalDecimalPlaces`;
- `ErrOverflow`;
- `ErrUnderflow`;
- `ErrNilDestination`;
- `ErrCBORSyntax`;
- `ErrCBORNonDeterministic`.

Use direct comparison or `errors.Is`. Do not parse error text and do not replace
typed failures with panic.

## 12. Common Errors by Agents

Do not:

- infer backend width from fractional decimal places;
- use one codec for values whose metadata resolves different scales;
- store a runtime scale on every numeric value or DB row;
- treat venue order tick/lot precision as token canonical scale;
- trim or case-normalize decimal input before parsing;
- silently truncate excess fractional digits;
- convert through `float64`, `shopspring/decimal`, or `math/big` when the
  approved contract already fits Sailfish units;
- call `String()` repeatedly in a hot loop when `AppendTo` or retained text is
  appropriate;
- use reflective CBOR for a measured hot positional record when direct append
  and prefix decode are available;
- expose naked units without the domain identity needed to interpret them;
- add aliases for removed APIs, fallback parsers, or old wire decoders.

## 13. Public API Map

| Concern | Public API |
|---|---|
| Static decimal-place policies | `DecimalPlaces0` through `DecimalPlaces20`, `StaticDecimalPlaces` |
| Unit constraints/providers | `Unit`, `NativeUnit`, `Uint8Units`, `Uint16Units`, `Uint32Units`, `Uint64Units`, `Uint256Units` |
| Semantic formats | `PriceInUint*Units`, `AmountInUint*Units`, `FixedDecimalFormat` |
| Value construction | `NewFixedDecimal`, `NewCompactFixedDecimal`, `NewFixedDecimalFromBytes`, `NewFixedDecimalFromUnits` |
| Static codec construction | `NewFixedDecimalCodec` |
| Runtime codec construction | `NewUint256FixedDecimalCodec` |
| Numeric state | `Units`, `SetUnits`, `IsZero` |
| Text-cache state | `HasRepresentation`, `Canonical` |
| Lengths | `Len`, `UnitsLen`, `CBORLen`, `MaxCBORSize` |
| Text parsing | `Parse`, `ParseCompact`, `ParseBytes`, `ParseUnits`, `ParseUnitsBytes`, runtime `ParseInto` forms |
| Text output | `AppendTo`, `AppendUnits`, `String`, `AppendText`, `MarshalText`, `UnmarshalText` |
| JSON | `AppendJSON`, `MarshalJSON`, `UnmarshalJSON` |
| CBOR | `AppendCBOR`, `MarshalCBOR`, `UnmarshalCBOR`, `ParseCBOR`, `ParseCBORFirst`, runtime `Into` forms |
| Same-format comparison | `Compare`, `Cmp`, `Equal`, `Less` methods |
| Cross-format comparison | package-level `Compare` |
| Arithmetic | `Add`, `AddAssign`, `AddOverflow`, `Sub`, `SubAssign`, `SubUnderflow` |

`Unit` is a closed set. Custom semantic formats embed one of the exported unit
providers and implement `FractionalDecimalPlaces()` on a zero-sized value type;
do not create another numeric backend through an interface wrapper.

`MaxCBORSize` is the maximum size of one Sailfish scalar, not an enclosing
record bound. The parent schema must include its array header, non-decimal
fields, and bounded string lengths.

## 14. Agent Checklist

Before using Sailfish, answer all of these explicitly:

1. Is the value unsigned?
2. Who owns its semantic kind: price, amount, or another domain type?
3. Where do fractional decimal places come from?
4. Are they static or metadata-driven?
5. Which backend covers the complete scaled-unit range?
6. Where is input parsed exactly once?
7. Does business logic operate only on units after that boundary?
8. Is output text retained, appended, or owned, and why?
9. Does CBOR parent schema supply scale and semantic identity out of band?
10. Are JSON decimals quoted strings?
11. Are overflow, underflow, precision, and malformed wire values fail-closed?
12. Is any claimed hot-path behavior covered by correctness, allocation, and
    benchmark evidence?

If any answer is unknown, stop at the domain contract. Do not guess a scale,
rounding rule, ownership model, or compatibility policy.
