package sailfish_test

import (
	"testing"

	"github.com/JekaMas/sailfish"
	"github.com/holiman/uint256"
)

type amountScale18 struct{ sailfish.Uint256Units }

func (amountScale18) NotionScale() sailfish.Notion { return 18 }

func TestPublicAPITypeInference(t *testing.T) {
	t.Parallel()

	price, err := sailfish.New[sailfish.PriceScale5]("123.31232")
	if err != nil {
		t.Fatal(err)
	}
	acceptPriceScale5(price)
	if price.Units() != 12_331_232 {
		t.Fatal(price.Units())
	}

	codec := sailfish.MustCodec[amountScale18]()
	amount, err := codec.Parse("1.000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	acceptAmountScale18(amount)
	if amount.Units() != (uint256.Int{1_000_000_000_000_000_001}) {
		t.Fatal(amount.Units())
	}
}

func acceptPriceScale5(sailfish.Decimal[sailfish.PriceScale5, uint64]) {}

func acceptAmountScale18(sailfish.Decimal[amountScale18, uint256.Int]) {}
