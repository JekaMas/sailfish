package sailfish_test

import (
	"fmt"

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
