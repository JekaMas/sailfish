package sailfish_test

import (
	"fmt"

	"github.com/JekaMas/sailfish"
)

func Example() {
	codec := sailfish.MustCodec[sailfish.PriceScale5]()
	price, _ := codec.Parse("123.31232")
	delta, _ := codec.Parse("0.00001")

	_ = price.AddAssign(delta)

	dst := make([]byte, 0, 32)
	dst = codec.AppendTo(dst, price)
	fmt.Println(string(dst))
	// Output: 123.31233
}
