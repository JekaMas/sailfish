package sailfish

import "github.com/holiman/uint256"

type uint64Scale0 struct{ Uint64Units }

func (uint64Scale0) NotionScale() Notion { return 0 }

type uint64Scale19 struct{ Uint64Units }

func (uint64Scale19) NotionScale() Notion { return 19 }

type uint64Scale20 struct{ Uint64Units }

func (uint64Scale20) NotionScale() Notion { return 20 }

type uint256Scale0 struct{ Uint256Units }

func (uint256Scale0) NotionScale() Notion { return 0 }

type uint256Scale6 struct{ Uint256Units }

func (uint256Scale6) NotionScale() Notion { return 6 }

type uint256Scale18 struct{ Uint256Units }

func (uint256Scale18) NotionScale() Notion { return 18 }

type uint256Scale37 struct{ Uint256Units }

func (uint256Scale37) NotionScale() Notion { return 37 }

type uint256Scale77 struct{ Uint256Units }

func (uint256Scale77) NotionScale() Notion { return 77 }

type uint256Scale78 struct{ Uint256Units }

func (uint256Scale78) NotionScale() Notion { return 78 }

type price5 = Decimal[PriceUint64[Fraction5], uint64]
type wide18 = Decimal[uint256Scale18, uint256.Int]

const maxUint256Decimal = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
const maxUint256PlusOne = "115792089237316195423570985008687907853269984665640564039457584007913129639936"
