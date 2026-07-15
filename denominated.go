package sailfish

// Denominated binds a fixed decimal to a comparable runtime identity. D can
// be a token identifier, a chain/token pair, or a base/quote market key. The
// fixed-decimal format continues to own fractional decimal places and unit
// representation; denomination is never used to infer or normalize scale.
type Denominated[D comparable, V FixedDecimalFormat[U], U Unit] struct {
	decimal      FixedDecimal[V, U]
	denomination D
}

// NewDenominated binds value to denomination without changing either value.
func NewDenominated[D comparable, V FixedDecimalFormat[U], U Unit](
	denomination D,
	value FixedDecimal[V, U],
) Denominated[D, V, U] {
	return Denominated[D, V, U]{decimal: value, denomination: denomination}
}

// Denomination returns the runtime identity associated with the decimal.
func (d Denominated[D, V, U]) Denomination() D { return d.denomination }

// Decimal returns the fixed decimal by value.
func (d Denominated[D, V, U]) Decimal() FixedDecimal[V, U] { return d.decimal }

// Compare compares values only when their denominations match.
func (d Denominated[D, V, U]) Compare(other Denominated[D, V, U]) (int, error) {
	if d.denomination != other.denomination {
		return 0, boxedErrDenominationMismatch
	}
	return d.decimal.Compare(other.decimal), nil
}

// Add adds values only when their denominations match.
func (d Denominated[D, V, U]) Add(
	other Denominated[D, V, U],
) (Denominated[D, V, U], error) {
	if d.denomination != other.denomination {
		return Denominated[D, V, U]{}, boxedErrDenominationMismatch
	}
	value, err := d.decimal.Add(other.decimal)
	if err != nil {
		return Denominated[D, V, U]{}, err
	}
	return Denominated[D, V, U]{decimal: value, denomination: d.denomination}, nil
}

// Sub subtracts values only when their denominations match.
func (d Denominated[D, V, U]) Sub(
	other Denominated[D, V, U],
) (Denominated[D, V, U], error) {
	if d.denomination != other.denomination {
		return Denominated[D, V, U]{}, boxedErrDenominationMismatch
	}
	value, err := d.decimal.Sub(other.decimal)
	if err != nil {
		return Denominated[D, V, U]{}, err
	}
	return Denominated[D, V, U]{decimal: value, denomination: d.denomination}, nil
}

// RescaleDenominated exactly rescales a value while preserving its runtime
// denomination. It never derives fractional decimal places from denomination.
func RescaleDenominated[
	ToV FixedDecimalFormat[ToU], ToU Unit,
	D comparable,
	FromV FixedDecimalFormat[FromU], FromU Unit,
](value Denominated[D, FromV, FromU]) (Denominated[D, ToV, ToU], error) {
	decimal, err := Rescale[ToV](value.decimal)
	if err != nil {
		return Denominated[D, ToV, ToU]{}, err
	}
	return Denominated[D, ToV, ToU]{
		decimal:      decimal,
		denomination: value.denomination,
	}, nil
}

// AddDenominatedAs checks runtime denomination, exactly rescales both values
// to the selected result format, and adds them.
func AddDenominatedAs[
	ResultV FixedDecimalFormat[ResultU], ResultU Unit,
	D comparable,
	AV FixedDecimalFormat[AU], AU Unit,
	BV FixedDecimalFormat[BU], BU Unit,
](
	a Denominated[D, AV, AU],
	b Denominated[D, BV, BU],
) (Denominated[D, ResultV, ResultU], error) {
	if a.denomination != b.denomination {
		return Denominated[D, ResultV, ResultU]{}, boxedErrDenominationMismatch
	}
	decimal, err := AddAs[ResultV](a.decimal, b.decimal)
	if err != nil {
		return Denominated[D, ResultV, ResultU]{}, err
	}
	return Denominated[D, ResultV, ResultU]{
		decimal:      decimal,
		denomination: a.denomination,
	}, nil
}

// SubDenominatedAs checks runtime denomination, exactly rescales both values
// to the selected result format, and subtracts them.
func SubDenominatedAs[
	ResultV FixedDecimalFormat[ResultU], ResultU Unit,
	D comparable,
	AV FixedDecimalFormat[AU], AU Unit,
	BV FixedDecimalFormat[BU], BU Unit,
](
	a Denominated[D, AV, AU],
	b Denominated[D, BV, BU],
) (Denominated[D, ResultV, ResultU], error) {
	if a.denomination != b.denomination {
		return Denominated[D, ResultV, ResultU]{}, boxedErrDenominationMismatch
	}
	decimal, err := SubAs[ResultV](a.decimal, b.decimal)
	if err != nil {
		return Denominated[D, ResultV, ResultU]{}, err
	}
	return Denominated[D, ResultV, ResultU]{
		decimal:      decimal,
		denomination: a.denomination,
	}, nil
}
