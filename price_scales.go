package sailfish

// PriceScale1 through PriceScale9 are ready-to-use uint64 price venues. A
// caller that needs a wider range or more precision can define a custom venue
// by embedding Uint256Units.
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
