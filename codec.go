package sailfish

// FixedDecimalCodec validates a fixed-decimal format once and carries its
// fractional decimal places through repeated parse and format operations. It
// is the preferred hot-loop API. Its zero value derives the decimal places
// from the compile-time format; NewFixedDecimalCodec validates and caches them.
//
// The one-byte decimalPlacesPlusOne encoding reserves zero for zero-value
// derivation.
type FixedDecimalCodec[V FixedDecimalFormat[U], U Unit] struct {
	decimalPlacesPlusOne uint8
}

// NewFixedDecimalCodec validates the format's fractional decimal places and
// caches them for repeated operations.
func NewFixedDecimalCodec[V FixedDecimalFormat[U], U Unit]() (FixedDecimalCodec[V, U], error) {
	decimalPlaces, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return FixedDecimalCodec[V, U]{}, boxedError(err)
	}
	return FixedDecimalCodec[V, U]{decimalPlacesPlusOne: uint8(decimalPlaces + 1)}, nil
}

func (c FixedDecimalCodec[V, U]) fractionalDecimalPlaces() int {
	if c.decimalPlacesPlusOne == 0 {
		decimalPlaces, _ := checkedFractionalDecimalPlaces[V, U]()
		return decimalPlaces
	}
	return int(c.decimalPlacesPlusOne - 1)
}

// FractionalDecimalPlaces returns the exact number of digits represented
// after the decimal point.
func (c FixedDecimalCodec[V, U]) FractionalDecimalPlaces() DecimalPlaces {
	return DecimalPlaces(c.fractionalDecimalPlaces())
}

// MaxIntegerDigits reports how many decimal digits can occur before the point
// in this backend's maximum value for the configured fractional decimal
// places. It describes capacity independently from fractional precision; it
// does not imply every value with that many digits fits the binary backend.
func (c FixedDecimalCodec[V, U]) MaxIntegerDigits() int {
	return unitDecimalDigits[U]() - c.fractionalDecimalPlaces()
}

// Parse retains s only when it is already canonical fixed-decimal text.
func (c FixedDecimalCodec[V, U]) Parse(s string) (FixedDecimal[V, U], error) {
	var format V
	units, canonical, err := format.unitParseString(s, c.fractionalDecimalPlaces())
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	d := FixedDecimal[V, U]{units: units}
	if canonical {
		d.representation = s
	}
	return d, nil
}

// ParseCompact never retains s.
func (c FixedDecimalCodec[V, U]) ParseCompact(s string) (FixedDecimal[V, U], error) {
	var format V
	units, _, err := format.unitParseString(s, c.fractionalDecimalPlaces())
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	return FixedDecimal[V, U]{units: units}, nil
}

// ParseBytes parses b directly and never retains it.
func (c FixedDecimalCodec[V, U]) ParseBytes(b []byte) (FixedDecimal[V, U], error) {
	var format V
	units, _, err := format.unitParseBytes(b, c.fractionalDecimalPlaces())
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	return FixedDecimal[V, U]{units: units}, nil
}

// ParseUnits parses strict fixed-decimal text directly into the selected unit
// backend. Use it when a numeric batch stores raw units for the smallest cache
// footprint and does not need FixedDecimal's optional retained representation.
// Successful and rejected parses allocate no memory.
func (c FixedDecimalCodec[V, U]) ParseUnits(s string) (U, Error) {
	var format V
	units, _, err := format.unitParseString(s, c.fractionalDecimalPlaces())
	return units, err
}

// ParseUnitsBytes is ParseUnits for byte input. It neither converts nor
// retains b.
func (c FixedDecimalCodec[V, U]) ParseUnitsBytes(b []byte) (U, Error) {
	var format V
	units, _, err := format.unitParseBytes(b, c.fractionalDecimalPlaces())
	return units, err
}

func (c FixedDecimalCodec[V, U]) FromUnits(units U) FixedDecimal[V, U] {
	_ = c.fractionalDecimalPlaces()
	return FixedDecimal[V, U]{units: units}
}

// UnitsLen returns the exact canonical text length of raw integer units.
func (c FixedDecimalCodec[V, U]) UnitsLen(units U) int {
	var format V
	return format.unitLen(units, c.fractionalDecimalPlaces())
}

// AppendUnits appends canonical fixed-decimal text directly from raw integer
// units. It allocates only when dst has insufficient capacity.
func (c FixedDecimalCodec[V, U]) AppendUnits(dst []byte, units U) []byte {
	var format V
	return format.unitAppend(dst, units, c.fractionalDecimalPlaces())
}

func (c FixedDecimalCodec[V, U]) Len(d FixedDecimal[V, U]) int {
	if d.representation != "" {
		return len(d.representation)
	}
	var format V
	return format.unitLen(d.units, c.fractionalDecimalPlaces())
}

func (c FixedDecimalCodec[V, U]) AppendTo(dst []byte, d FixedDecimal[V, U]) []byte {
	if d.representation != "" {
		return append(dst, d.representation...)
	}
	var format V
	return format.unitAppend(dst, d.units, c.fractionalDecimalPlaces())
}

func (c FixedDecimalCodec[V, U]) AppendJSON(dst []byte, d FixedDecimal[V, U]) []byte {
	dst = append(dst, '"')
	dst = c.AppendTo(dst, d)
	return append(dst, '"')
}

func (c FixedDecimalCodec[V, U]) String(d FixedDecimal[V, U]) string {
	if d.representation != "" {
		return d.representation
	}
	var format V
	return format.unitString(d.units, c.fractionalDecimalPlaces())
}

func (c FixedDecimalCodec[V, U]) Canonical(d FixedDecimal[V, U]) FixedDecimal[V, U] {
	if d.representation == "" {
		d.representation = c.String(d)
	}
	return d
}
