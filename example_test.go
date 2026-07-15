package sailfish_test

import (
	"fmt"
	"math/big"

	"github.com/JekaMas/sailfish"
	"github.com/fxamacker/cbor/v2"
	json "github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

type examplePriceFormat = sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]
type examplePrice = sailfish.FixedDecimal[examplePriceFormat, uint64]
type exampleAmountFormat = sailfish.AmountInUint256Units[sailfish.DecimalPlaces18]
type exampleAmount = sailfish.FixedDecimal[exampleAmountFormat, uint256.Int]

func ExampleFixedDecimalCodec_price() {
	codec, err := sailfish.NewFixedDecimalCodec[examplePriceFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}

	price, err := codec.Parse("123.31232")
	if err != nil {
		fmt.Println(err)
		return
	}
	next, err := price.Add(codec.FromUnits(1))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(next.String())
	fmt.Println(next.Units())
	// Output:
	// 123.31233
	// 12331233
}

func ExampleNewFixedDecimalFromUnits() {
	type AmountFormat = sailfish.AmountInUint32Units[sailfish.DecimalPlaces6]

	amount, err := sailfish.NewFixedDecimalFromUnits[AmountFormat](uint32(1_234_567))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(amount.String())
	// Output:
	// 1.234567
}

func ExampleFixedDecimal_integerConversions() {
	codec, err := sailfish.NewFixedDecimalCodec[exampleAmountFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Integer conversions use already-scaled units. They do not rescale.
	value, err := codec.FromU256(uint256.Int{1_250_000_000_000_000_000})
	if err != nil {
		fmt.Println(err)
		return
	}
	var destination big.Int
	if err = value.ToBigInt(&destination); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(value.String())
	fmt.Println(destination.String())
	// Output:
	// 1.250000000000000000
	// 1250000000000000000
}

func ExampleFixedDecimal_rationalConversions() {
	codec, err := sailfish.NewFixedDecimalCodec[examplePriceFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}

	price, err := codec.FromBigRat(big.NewRat(385_351, 3_125))
	if err != nil {
		fmt.Println(err)
		return
	}
	var rational big.Rat
	var workspace sailfish.BigRatWorkspace
	if err = price.ToBigRat(&rational, &workspace); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(price.String())
	fmt.Println(rational.String())
	// Output:
	// 123.31232
	// 385351/3125
}

func ExampleAddDenominatedAs() {
	type Asset struct {
		Chain uint32
		Token string
	}
	type Price2 = sailfish.PriceInUint64Units[sailfish.DecimalPlaces2]
	type Price5 = sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]

	asset := Asset{Chain: 1, Token: "USDC"}
	price2, _ := sailfish.NewFixedDecimal[Price2]("1.20")
	price5, _ := sailfish.NewFixedDecimal[Price5]("0.00003")
	left := sailfish.NewDenominated(asset, price2)
	right := sailfish.NewDenominated(asset, price5)

	sum, err := sailfish.AddDenominatedAs[Price5](left, right)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(sum.Denomination().Token)
	fmt.Println(sum.Decimal().String())
	// Output:
	// USDC
	// 1.20003
}

func ExampleUint256FixedDecimalCodec() {
	codec, err := sailfish.NewUint256FixedDecimalCodec(18)
	if err != nil {
		fmt.Println(err)
		return
	}

	var units uint256.Int
	if parseErr := codec.ParseInto("1.250000000000000000", &units); parseErr != "" {
		fmt.Println(parseErr)
		return
	}

	fmt.Println(string(codec.AppendTo(nil, units)))
	// Output:
	// 1.250000000000000000
}

type exampleQuote struct {
	_ struct{} `cbor:",toarray"`

	Price  examplePrice
	Amount exampleAmount
}

func ExampleFixedDecimal_serialization() {
	priceCodec, err := sailfish.NewFixedDecimalCodec[examplePriceFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}
	amountCodec, err := sailfish.NewFixedDecimalCodec[exampleAmountFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}
	price, err := priceCodec.Parse("123.31232")
	if err != nil {
		fmt.Println(err)
		return
	}
	amount, err := amountCodec.Parse("1.250000000000000000")
	if err != nil {
		fmt.Println(err)
		return
	}
	quote := exampleQuote{Price: price, Amount: amount}

	jsonRaw, err := json.Marshal(quote.Price)
	if err != nil {
		fmt.Println(err)
		return
	}
	enc, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		fmt.Println(err)
		return
	}
	cborRaw, err := enc.Marshal(quote)
	if err != nil {
		fmt.Println(err)
		return
	}
	dec, err := cbor.DecOptions{}.DecMode()
	if err != nil {
		fmt.Println(err)
		return
	}
	var decoded exampleQuote
	if err := dec.Unmarshal(cborRaw, &decoded); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(jsonRaw))
	fmt.Println(decoded.Price.String())
	fmt.Println(decoded.Amount.String())
	// Output:
	// "123.31232"
	// 123.31232
	// 1.250000000000000000
}

func ExampleFixedDecimalCodec_manualPositionalCBOR() {
	priceCodec, err := sailfish.NewFixedDecimalCodec[examplePriceFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}
	amountCodec, err := sailfish.NewFixedDecimalCodec[exampleAmountFormat]()
	if err != nil {
		fmt.Println(err)
		return
	}
	price := priceCodec.FromUnits(12_331_232)
	var amountUnits uint256.Int
	amountUnits.SetUint64(1_250_000_000_000_000_000)
	amount := amountCodec.FromUnits(amountUnits)

	record := make([]byte, 0, 1+2*sailfish.MaxCBORSize)
	record = append(record, 0x82) // fixed two-field CBOR array
	record = priceCodec.AppendCBOR(record, price)
	record = amountCodec.AppendCBOR(record, amount)

	raw := record[1:]
	decodedPrice, raw, err := priceCodec.ParseCBORFirst(raw)
	if err != nil {
		fmt.Println(err)
		return
	}
	decodedAmount, raw, err := amountCodec.ParseCBORFirst(raw)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(decodedPrice.String())
	fmt.Println(decodedAmount.String())
	fmt.Println(len(raw))
	// Output:
	// 123.31232
	// 1.250000000000000000
	// 0
}
