package sailfish_test

import (
	"fmt"

	"github.com/JekaMas/sailfish"
	"github.com/holiman/uint256"
)

func Example() {
	codec, err := sailfish.NewCodec[sailfish.PriceUint64[sailfish.Fraction5]]()
	if err != nil {
		return
	}
	price, _ := codec.Parse("123.31232")
	delta, _ := codec.Parse("0.00001")

	_ = price.AddAssign(delta)

	dst := make([]byte, 0, 32)
	dst = codec.AppendTo(dst, price)
	fmt.Println(string(dst))
	// Output: 123.31233
}

func ExampleUint256Codec() {
	codec, err := sailfish.NewUint256Codec(6)
	if err != nil {
		return
	}
	var units uint256.Int
	if parseErr := codec.ParseInto("123.456789", &units); parseErr != "" {
		return
	}

	dst := codec.AppendTo(make([]byte, 0, 32), units)
	fmt.Println(string(dst))
	// Output: 123.456789
}
