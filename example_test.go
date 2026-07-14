package sailfish_test

import (
	"fmt"

	"github.com/JekaMas/sailfish"
	"github.com/holiman/uint256"
)

func Example() {
	codec := sailfish.MustCodec[sailfish.PriceUint64[sailfish.Fraction5]]()
	price, _ := codec.Parse("123.31232")
	delta, _ := codec.Parse("0.00001")

	_ = price.AddAssign(delta)

	dst := make([]byte, 0, 32)
	dst = codec.AppendTo(dst, price)
	fmt.Println(string(dst))
	// Output: 123.31233
}

func ExampleUint256Codec() {
	codec := sailfish.MustUint256Codec(6)
	var units uint256.Int
	if err := codec.ParseInto("123.456789", &units); err != "" {
		panic(err)
	}

	dst := codec.AppendTo(make([]byte, 0, 32), units)
	fmt.Println(string(dst))
	// Output: 123.456789
}
