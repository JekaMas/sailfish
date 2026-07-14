package sailfish

// Codec validates a venue once and carries its scale through repeated
// parse/format operations. It is the preferred hot-loop API. Its zero value is
// usable and derives scale from the compile-time venue; NewCodec validates and
// caches that scale for the hot path.
//
// The one-byte scalePlusOne encoding reserves zero for zero-value derivation.
type Codec[V Venue[U], U Unit] struct {
	scalePlusOne uint8
}

func NewCodec[V Venue[U], U Unit]() (Codec[V, U], error) {
	scale, err := checkedScale[V, U]()
	if err != "" {
		return Codec[V, U]{}, boxedError(err)
	}
	return Codec[V, U]{scalePlusOne: uint8(scale + 1)}, nil
}

func (c Codec[V, U]) scale() int {
	if c.scalePlusOne == 0 {
		scale, _ := checkedScale[V, U]()
		return scale
	}
	return int(c.scalePlusOne - 1)
}

func (c Codec[V, U]) Scale() Notion { return Notion(c.scale()) }

// MaxIntegerDigits reports how many decimal digits can occur before the point
// in this backend's maximum value at the configured scale. It describes
// capacity independently from fractional scale; it does not imply that every
// value with that many digits fits the binary backend.
func (c Codec[V, U]) MaxIntegerDigits() int {
	return unitDecimalDigits[U]() - c.scale()
}

// Parse retains s only when it is already canonical fixed-scale text.
func (c Codec[V, U]) Parse(s string) (Decimal[V, U], error) {
	var venue V
	units, canonical, err := venue.unitParseString(s, c.scale())
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	d := Decimal[V, U]{units: units}
	if canonical {
		d.representation = s
	}
	return d, nil
}

// ParseCompact never retains s.
func (c Codec[V, U]) ParseCompact(s string) (Decimal[V, U], error) {
	var venue V
	units, _, err := venue.unitParseString(s, c.scale())
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	return Decimal[V, U]{units: units}, nil
}

// ParseBytes parses b directly and never retains it.
func (c Codec[V, U]) ParseBytes(b []byte) (Decimal[V, U], error) {
	var venue V
	units, _, err := venue.unitParseBytes(b, c.scale())
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	return Decimal[V, U]{units: units}, nil
}

func (c Codec[V, U]) FromUnits(units U) Decimal[V, U] {
	_ = c.scale()
	return Decimal[V, U]{units: units}
}

func (c Codec[V, U]) Len(d Decimal[V, U]) int {
	if d.representation != "" {
		return len(d.representation)
	}
	var venue V
	return venue.unitLen(d.units, c.scale())
}

func (c Codec[V, U]) AppendTo(dst []byte, d Decimal[V, U]) []byte {
	if d.representation != "" {
		return append(dst, d.representation...)
	}
	var venue V
	return venue.unitAppend(dst, d.units, c.scale())
}

func (c Codec[V, U]) AppendJSON(dst []byte, d Decimal[V, U]) []byte {
	dst = append(dst, '"')
	dst = c.AppendTo(dst, d)
	return append(dst, '"')
}

func (c Codec[V, U]) String(d Decimal[V, U]) string {
	if d.representation != "" {
		return d.representation
	}
	var venue V
	return venue.unitString(d.units, c.scale())
}

func (c Codec[V, U]) Canonical(d Decimal[V, U]) Decimal[V, U] {
	if d.representation == "" {
		d.representation = c.String(d)
	}
	return d
}
