package sailfish

// DecimalPlaces0 through DecimalPlaces20 are zero-sized policies that state
// the exact number of digits represented after the decimal point. Decimal
// places are independent from the scaled-integer backend: callers choose both
// the fractional precision and numeric capacity.
type DecimalPlaces0 struct{}

func (DecimalPlaces0) FractionalDecimalPlaces() DecimalPlaces { return 0 }

type DecimalPlaces1 struct{}

func (DecimalPlaces1) FractionalDecimalPlaces() DecimalPlaces { return 1 }

type DecimalPlaces2 struct{}

func (DecimalPlaces2) FractionalDecimalPlaces() DecimalPlaces { return 2 }

type DecimalPlaces3 struct{}

func (DecimalPlaces3) FractionalDecimalPlaces() DecimalPlaces { return 3 }

type DecimalPlaces4 struct{}

func (DecimalPlaces4) FractionalDecimalPlaces() DecimalPlaces { return 4 }

type DecimalPlaces5 struct{}

func (DecimalPlaces5) FractionalDecimalPlaces() DecimalPlaces { return 5 }

type DecimalPlaces6 struct{}

func (DecimalPlaces6) FractionalDecimalPlaces() DecimalPlaces { return 6 }

type DecimalPlaces7 struct{}

func (DecimalPlaces7) FractionalDecimalPlaces() DecimalPlaces { return 7 }

type DecimalPlaces8 struct{}

func (DecimalPlaces8) FractionalDecimalPlaces() DecimalPlaces { return 8 }

type DecimalPlaces9 struct{}

func (DecimalPlaces9) FractionalDecimalPlaces() DecimalPlaces { return 9 }

type DecimalPlaces10 struct{}

func (DecimalPlaces10) FractionalDecimalPlaces() DecimalPlaces { return 10 }

type DecimalPlaces11 struct{}

func (DecimalPlaces11) FractionalDecimalPlaces() DecimalPlaces { return 11 }

type DecimalPlaces12 struct{}

func (DecimalPlaces12) FractionalDecimalPlaces() DecimalPlaces { return 12 }

type DecimalPlaces13 struct{}

func (DecimalPlaces13) FractionalDecimalPlaces() DecimalPlaces { return 13 }

type DecimalPlaces14 struct{}

func (DecimalPlaces14) FractionalDecimalPlaces() DecimalPlaces { return 14 }

type DecimalPlaces15 struct{}

func (DecimalPlaces15) FractionalDecimalPlaces() DecimalPlaces { return 15 }

type DecimalPlaces16 struct{}

func (DecimalPlaces16) FractionalDecimalPlaces() DecimalPlaces { return 16 }

type DecimalPlaces17 struct{}

func (DecimalPlaces17) FractionalDecimalPlaces() DecimalPlaces { return 17 }

type DecimalPlaces18 struct{}

func (DecimalPlaces18) FractionalDecimalPlaces() DecimalPlaces { return 18 }

type DecimalPlaces19 struct{}

func (DecimalPlaces19) FractionalDecimalPlaces() DecimalPlaces { return 19 }

type DecimalPlaces20 struct{}

func (DecimalPlaces20) FractionalDecimalPlaces() DecimalPlaces { return 20 }

func fractionalDecimalPlaces[S StaticDecimalPlaces]() DecimalPlaces {
	var decimalPlaces S
	return decimalPlaces.FractionalDecimalPlaces()
}

// PriceInUint8Units through PriceInUint256Units identify prices represented by
// one unsigned scaled integer of the named width. The type parameter states
// the exact fractional decimal places. For example,
// PriceInUint64Units[DecimalPlaces5] stores uint64 units and represents
// numeric value as units / 100000. Backend width controls range; it is not
// inferred from the decimal places.
type PriceInUint8Units[S StaticDecimalPlaces] struct {
	Uint8Units
}

func (PriceInUint8Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type PriceInUint16Units[S StaticDecimalPlaces] struct {
	Uint16Units
}

func (PriceInUint16Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type PriceInUint32Units[S StaticDecimalPlaces] struct {
	Uint32Units
}

func (PriceInUint32Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type PriceInUint64Units[S StaticDecimalPlaces] struct {
	Uint64Units
}

func (PriceInUint64Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type PriceInUint256Units[S StaticDecimalPlaces] struct {
	Uint256Units
}

func (PriceInUint256Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

// AmountInUint8Units through AmountInUint256Units are the amount-kind
// equivalents. Price and amount formats remain distinct types even with equal
// fractional decimal places and the same backend.
type AmountInUint8Units[S StaticDecimalPlaces] struct {
	Uint8Units
}

func (AmountInUint8Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type AmountInUint16Units[S StaticDecimalPlaces] struct {
	Uint16Units
}

func (AmountInUint16Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type AmountInUint32Units[S StaticDecimalPlaces] struct {
	Uint32Units
}

func (AmountInUint32Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type AmountInUint64Units[S StaticDecimalPlaces] struct {
	Uint64Units
}

func (AmountInUint64Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}

type AmountInUint256Units[S StaticDecimalPlaces] struct {
	Uint256Units
}

func (AmountInUint256Units[S]) FractionalDecimalPlaces() DecimalPlaces {
	return fractionalDecimalPlaces[S]()
}
