package sailfish

// Fraction0 through Fraction20 are zero-sized fractional-scale policies.
// Scale is independent from the scaled-integer backend: callers choose both
// the number of digits after the point and the required numeric capacity.
type Fraction0 struct{}

func (Fraction0) NotionScale() Notion { return 0 }

type Fraction1 struct{}

func (Fraction1) NotionScale() Notion { return 1 }

type Fraction2 struct{}

func (Fraction2) NotionScale() Notion { return 2 }

type Fraction3 struct{}

func (Fraction3) NotionScale() Notion { return 3 }

type Fraction4 struct{}

func (Fraction4) NotionScale() Notion { return 4 }

type Fraction5 struct{}

func (Fraction5) NotionScale() Notion { return 5 }

type Fraction6 struct{}

func (Fraction6) NotionScale() Notion { return 6 }

type Fraction7 struct{}

func (Fraction7) NotionScale() Notion { return 7 }

type Fraction8 struct{}

func (Fraction8) NotionScale() Notion { return 8 }

type Fraction9 struct{}

func (Fraction9) NotionScale() Notion { return 9 }

type Fraction10 struct{}

func (Fraction10) NotionScale() Notion { return 10 }

type Fraction11 struct{}

func (Fraction11) NotionScale() Notion { return 11 }

type Fraction12 struct{}

func (Fraction12) NotionScale() Notion { return 12 }

type Fraction13 struct{}

func (Fraction13) NotionScale() Notion { return 13 }

type Fraction14 struct{}

func (Fraction14) NotionScale() Notion { return 14 }

type Fraction15 struct{}

func (Fraction15) NotionScale() Notion { return 15 }

type Fraction16 struct{}

func (Fraction16) NotionScale() Notion { return 16 }

type Fraction17 struct{}

func (Fraction17) NotionScale() Notion { return 17 }

type Fraction18 struct{}

func (Fraction18) NotionScale() Notion { return 18 }

type Fraction19 struct{}

func (Fraction19) NotionScale() Notion { return 19 }

type Fraction20 struct{}

func (Fraction20) NotionScale() Notion { return 20 }

func notionScale[S VenueScale]() Notion {
	var scale S
	return scale.NotionScale()
}

// PriceUint8 through PriceUint256 combine price identity and a fractional
// scale with an explicit scaled-integer backend. Backend width controls range;
// it is not inferred from fractional scale.
type PriceUint8[S VenueScale] struct {
	Uint8Units
}

func (PriceUint8[S]) NotionScale() Notion { return notionScale[S]() }

type PriceUint16[S VenueScale] struct {
	Uint16Units
}

func (PriceUint16[S]) NotionScale() Notion { return notionScale[S]() }

type PriceUint32[S VenueScale] struct {
	Uint32Units
}

func (PriceUint32[S]) NotionScale() Notion { return notionScale[S]() }

type PriceUint64[S VenueScale] struct {
	Uint64Units
}

func (PriceUint64[S]) NotionScale() Notion { return notionScale[S]() }

type PriceUint256[S VenueScale] struct {
	Uint256Units
}

func (PriceUint256[S]) NotionScale() Notion { return notionScale[S]() }

// AmountUint8 through AmountUint256 are the amount-kind equivalents. Price
// and amount formats remain distinct types even with equal scale and backend.
type AmountUint8[S VenueScale] struct {
	Uint8Units
}

func (AmountUint8[S]) NotionScale() Notion { return notionScale[S]() }

type AmountUint16[S VenueScale] struct {
	Uint16Units
}

func (AmountUint16[S]) NotionScale() Notion { return notionScale[S]() }

type AmountUint32[S VenueScale] struct {
	Uint32Units
}

func (AmountUint32[S]) NotionScale() Notion { return notionScale[S]() }

type AmountUint64[S VenueScale] struct {
	Uint64Units
}

func (AmountUint64[S]) NotionScale() Notion { return notionScale[S]() }

type AmountUint256[S VenueScale] struct {
	Uint256Units
}

func (AmountUint256[S]) NotionScale() Notion { return notionScale[S]() }

// PriceScale1 through PriceScale9 retain the original wide-range uint64
// defaults. Their names specify fractional digits only. They are concrete
// because direct New and Decimal methods resolve concrete scale metadata about
// one nanosecond faster than generic scale composition on the measured Go
// toolchain. Cached Codec operations are equivalent. Use PriceUint8,
// PriceUint16, PriceUint32, or PriceUint256 when a different range is desired.
type PriceScale1 struct{ Uint64Units }

func (PriceScale1) NotionScale() Notion { return 1 }

type PriceScale2 struct{ Uint64Units }

func (PriceScale2) NotionScale() Notion { return 2 }

type PriceScale3 struct{ Uint64Units }

func (PriceScale3) NotionScale() Notion { return 3 }

type PriceScale4 struct{ Uint64Units }

func (PriceScale4) NotionScale() Notion { return 4 }

type PriceScale5 struct{ Uint64Units }

func (PriceScale5) NotionScale() Notion { return 5 }

type PriceScale6 struct{ Uint64Units }

func (PriceScale6) NotionScale() Notion { return 6 }

type PriceScale7 struct{ Uint64Units }

func (PriceScale7) NotionScale() Notion { return 7 }

type PriceScale8 struct{ Uint64Units }

func (PriceScale8) NotionScale() Notion { return 8 }

type PriceScale9 struct{ Uint64Units }

func (PriceScale9) NotionScale() Notion { return 9 }

// AmountScale18 is the common on-chain amount format: 18 fractional digits
// backed by uint256.Int. It is concrete for the same direct-call reason as the
// built-in price formats. Other combinations use AmountUint*[FractionN].
type AmountScale18 struct{ Uint256Units }

func (AmountScale18) NotionScale() Notion { return 18 }
