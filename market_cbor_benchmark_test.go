package sailfish

import (
	_ "embed"
	"sort"
	"testing"
	"time"

	json "github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

//go:embed testdata/market_cbor_samples.json
var marketCBORSnapshotJSON []byte

type marketCBORSnapshot struct {
	SnapshotAt string             `json:"snapshot_at"`
	Sources    []marketCBORSource `json:"sources"`
}

type marketCBORSource struct {
	Venue      string             `json:"venue"`
	MarketType string             `json:"market_type"`
	Markets    []marketCBORMarket `json:"markets"`
}

type marketCBORMarket struct {
	Venue         string   `json:"venue"`
	MarketType    string   `json:"market_type"`
	Rank          int      `json:"rank"`
	Symbol        string   `json:"symbol"`
	PriceScale    Notion   `json:"price_scale"`
	QuantityScale Notion   `json:"quantity_scale"`
	Prices        []string `json:"prices"`
	Quantities    []string `json:"quantities"`
}

type preparedMarketCBORCohort struct {
	name       string
	prices     []preparedMarketCBORScalar
	quantities []preparedMarketCBORScalar
	bars       [3][]preparedMarketCBORBar
}

type preparedMarketCBORScalar struct {
	codec Uint256Codec
	units uint256.Int
	wire  []byte
}

type preparedMarketCBORBar struct {
	priceCodec    Uint256Codec
	quantityCodec Uint256Codec
	symbol        string
	prices        [4]uint256.Int
	volume        uint256.Int
	firstUpdateID uint64
	lastUpdateID  uint64
	updateCount   uint64
	barStartMS    uint64
	barCloseMS    uint64
	finalized     uint64
	flags         uint16
}

type marketCBORScalarKey struct {
	kind  uint8
	scale Notion
	units uint256.Int
}

const (
	marketCBORQuantityMin = iota
	marketCBORQuantityMedian
	marketCBORQuantityMax
	marketCBORQuantityModes
)

var (
	preparedMarketCBORBarSink      preparedMarketCBORBar
	marketCBORQuantityNames        = [marketCBORQuantityModes]string{"quantity_min", "quantity_median", "quantity_max"}
	marketCBORExpectedBarWireStats = map[string]marketCBORWireStats{
		"mexc_spot/quantity_min":           {minimum: 55, p50: 60, p95: 68, maximum: 75, total: 6096, count: 100},
		"mexc_spot/quantity_median":        {minimum: 59, p50: 63, p95: 71, maximum: 78, total: 6366, count: 100},
		"mexc_spot/quantity_max":           {minimum: 62, p50: 64, p95: 74, maximum: 78, total: 6592, count: 100},
		"hyperliquid_spot/quantity_min":    {minimum: 48, p50: 56, p95: 64, maximum: 69, total: 5742, count: 100},
		"hyperliquid_spot/quantity_median": {minimum: 48, p50: 60, p95: 66, maximum: 71, total: 6018, count: 100},
		"hyperliquid_spot/quantity_max":    {minimum: 48, p50: 60, p95: 68, maximum: 73, total: 6080, count: 100},
		"hyperliquid_perp/quantity_min":    {minimum: 53, p50: 56, p95: 64, maximum: 66, total: 5786, count: 100},
		"hyperliquid_perp/quantity_median": {minimum: 56, p50: 60, p95: 67, maximum: 68, total: 6061, count: 100},
		"hyperliquid_perp/quantity_max":    {minimum: 57, p50: 60, p95: 68, maximum: 69, total: 6152, count: 100},
	}
)

func TestCBORRealMarketSnapshot(t *testing.T) {
	cohorts := prepareMarketCBORCohorts(t)
	if len(cohorts) != 3 {
		t.Fatalf("cohort count = %d, want 3", len(cohorts))
	}
	if len(marketCBORExpectedBarWireStats) != len(cohorts)*marketCBORQuantityModes {
		t.Fatalf("expected wire distributions = %d, want %d", len(marketCBORExpectedBarWireStats), len(cohorts)*marketCBORQuantityModes)
	}

	for _, cohort := range cohorts {
		for mode, bars := range cohort.bars {
			name := cohort.name + "/" + marketCBORQuantityNames[mode]
			if len(bars) != 100 {
				t.Fatalf("%s markets = %d, want 100", name, len(bars))
			}
			stats := marketCBORBarWireStats(bars)
			expected, ok := marketCBORExpectedBarWireStats[name]
			if !ok {
				t.Fatalf("%s has no expected wire distribution", name)
			}
			if stats != expected {
				t.Fatalf("%s wire stats = %+v, want %+v", name, stats, expected)
			}
			for _, bar := range bars {
				wire := appendPreparedMarketCBORBar(make([]byte, 0, 128), bar)
				decoded, err := decodePreparedMarketCBORBar(wire, bar.priceCodec, bar.quantityCodec)
				if err != nil {
					t.Fatalf("%s/%s decode: %v", name, bar.symbol, err)
				}
				if decoded.symbol != bar.symbol || decoded.prices != bar.prices || decoded.volume != bar.volume {
					t.Fatalf("%s/%s round trip mismatch", name, bar.symbol)
				}
			}
		}
	}
}

func TestCBORManualPositionalBarTheoreticalMinimum(t *testing.T) {
	codec, err := NewUint256Codec(0)
	if err != nil {
		t.Fatal(err)
	}
	bar := preparedMarketCBORBar{priceCodec: codec, quantityCodec: codec}
	wire := appendPreparedMarketCBORBar(make([]byte, 0, 15), bar)
	if got := len(wire); got != 15 {
		t.Fatalf("minimum wire length = %d, want 15", got)
	}
	decoded, err := decodePreparedMarketCBORBar(wire, codec, codec)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != bar {
		t.Fatalf("minimum round trip = %#v, want %#v", decoded, bar)
	}
}

func BenchmarkCBORRealMarketScalars(b *testing.B) {
	for _, cohort := range prepareMarketCBORCohorts(b) {
		benchmarkMarketCBORScalars(b, cohort.name+"/price", cohort.prices)
		benchmarkMarketCBORScalars(b, cohort.name+"/quantity", cohort.quantities)
	}
}

func BenchmarkCBORRealMarketBars(b *testing.B) {
	for _, cohort := range prepareMarketCBORCohorts(b) {
		for mode, bars := range cohort.bars {
			name := cohort.name + "/" + marketCBORQuantityNames[mode]
			stats := marketCBORBarWireStats(bars)
			b.Run(name+"/encode", func(b *testing.B) {
				buffer := make([]byte, 0, stats.maximum)
				index := 0
				b.ReportAllocs()
				for b.Loop() {
					benchBytesSink = appendPreparedMarketCBORBar(buffer[:0], bars[index])
					index++
					if index == len(bars) {
						index = 0
					}
				}
				reportMarketCBORWireStats(b, stats)
			})

			wires := make([][]byte, len(bars))
			for i, bar := range bars {
				wires[i] = appendPreparedMarketCBORBar(make([]byte, 0, stats.maximum), bar)
			}
			b.Run(name+"/decode", func(b *testing.B) {
				index := 0
				b.ReportAllocs()
				for b.Loop() {
					bar := bars[index]
					preparedMarketCBORBarSink, _ = decodePreparedMarketCBORBar(
						wires[index], bar.priceCodec, bar.quantityCodec,
					)
					index++
					if index == len(bars) {
						index = 0
					}
				}
				reportMarketCBORWireStats(b, stats)
			})
		}
	}
}

func prepareMarketCBORCohorts(tb testing.TB) []preparedMarketCBORCohort {
	tb.Helper()
	var snapshot marketCBORSnapshot
	if err := json.Unmarshal(marketCBORSnapshotJSON, &snapshot); err != nil {
		tb.Fatal(err)
	}
	snapshotTime, err := time.Parse(time.RFC3339, snapshot.SnapshotAt)
	if err != nil {
		tb.Fatal(err)
	}
	barStartMS := uint64(snapshotTime.UnixMilli()/60_000) * 60_000

	identities := make(map[string]struct{}, 300)
	cohorts := make([]preparedMarketCBORCohort, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		if len(source.Markets) != 100 {
			tb.Fatalf("%s/%s market count = %d, want 100", source.Venue, source.MarketType, len(source.Markets))
		}
		cohort := preparedMarketCBORCohort{name: source.Venue + "_" + source.MarketType}
		scalarSeen := make(map[marketCBORScalarKey]struct{})
		for position, market := range source.Markets {
			if market.Venue != source.Venue || market.MarketType != source.MarketType || market.Rank != position+1 {
				tb.Fatalf("%s/%s rank %d has inconsistent identity", source.Venue, source.MarketType, position+1)
			}
			identity := market.Venue + "\x00" + market.MarketType + "\x00" + market.Symbol
			if _, duplicate := identities[identity]; duplicate {
				tb.Fatalf("duplicate market identity %q", identity)
			}
			identities[identity] = struct{}{}

			priceCodec, codecErr := NewUint256Codec(market.PriceScale)
			if codecErr != nil {
				tb.Fatalf("%s price scale: %v", market.Symbol, codecErr)
			}
			quantityCodec, codecErr := NewUint256Codec(market.QuantityScale)
			if codecErr != nil {
				tb.Fatalf("%s quantity scale: %v", market.Symbol, codecErr)
			}
			prices := prepareMarketCBORValues(tb, market.Symbol, "price", priceCodec, market.Prices)
			quantities := prepareMarketCBORValues(tb, market.Symbol, "quantity", quantityCodec, market.Quantities)
			appendUniqueMarketCBORScalars(&cohort.prices, scalarSeen, 0, priceCodec, prices)
			appendUniqueMarketCBORScalars(&cohort.quantities, scalarSeen, 1, quantityCodec, quantities)

			var barPrices [4]uint256.Int
			for i := range barPrices {
				barPrices[i] = prices[i*(len(prices)-1)/(len(barPrices)-1)]
			}
			quantityIndexes := [marketCBORQuantityModes]int{0, len(quantities) / 2, len(quantities) - 1}
			for mode, quantityIndex := range quantityIndexes {
				firstUpdateID := uint64(100_000 + market.Rank*100)
				cohort.bars[mode] = append(cohort.bars[mode], preparedMarketCBORBar{
					priceCodec:    priceCodec,
					quantityCodec: quantityCodec,
					symbol:        market.Symbol,
					prices:        barPrices,
					volume:        quantities[quantityIndex],
					firstUpdateID: firstUpdateID,
					lastUpdateID:  firstUpdateID + 63,
					updateCount:   64,
					barStartMS:    barStartMS,
					barCloseMS:    barStartMS + 60_000,
					finalized:     uint64(12_345_678 + market.Rank),
					flags:         1,
				})
			}
		}
		cohorts = append(cohorts, cohort)
	}
	if len(identities) != 300 {
		tb.Fatalf("unique market identities = %d, want 300", len(identities))
	}
	return cohorts
}

func prepareMarketCBORValues(
	tb testing.TB,
	symbol string,
	kind string,
	codec Uint256Codec,
	values []string,
) []uint256.Int {
	tb.Helper()
	if len(values) == 0 {
		tb.Fatalf("%s has no %s samples", symbol, kind)
	}
	seen := make(map[string]struct{}, len(values))
	units := make([]uint256.Int, 0, len(values))
	for _, value := range values {
		if _, duplicate := seen[value]; duplicate {
			tb.Fatalf("%s has duplicate %s sample %q", symbol, kind, value)
		}
		seen[value] = struct{}{}
		parsed, parseErr := codec.Parse(value)
		if parseErr != "" {
			tb.Fatalf("%s %s %q: %v", symbol, kind, value, parseErr)
		}
		if parsed.IsZero() {
			tb.Fatalf("%s has zero %s sample %q", symbol, kind, value)
		}
		if len(units) > 0 && compareUint256(units[len(units)-1], parsed) >= 0 {
			tb.Fatalf("%s %s samples are not strictly increasing", symbol, kind)
		}
		units = append(units, parsed)
	}
	return units
}

func appendUniqueMarketCBORScalars(
	destination *[]preparedMarketCBORScalar,
	seen map[marketCBORScalarKey]struct{},
	kind uint8,
	codec Uint256Codec,
	values []uint256.Int,
) {
	for _, units := range values {
		key := marketCBORScalarKey{kind: kind, scale: codec.Scale(), units: units}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		wire := codec.AppendCBOR(make([]byte, 0, MaxCBORSize), units)
		*destination = append(*destination, preparedMarketCBORScalar{codec: codec, units: units, wire: wire})
	}
}

func benchmarkMarketCBORScalars(b *testing.B, name string, values []preparedMarketCBORScalar) {
	stats := marketCBORScalarWireStats(values)
	b.Run(name+"/encode", func(b *testing.B) {
		buffer := make([]byte, 0, MaxCBORSize)
		index := 0
		b.ReportAllocs()
		for b.Loop() {
			value := values[index]
			benchBytesSink = value.codec.AppendCBOR(buffer[:0], value.units)
			index++
			if index == len(values) {
				index = 0
			}
		}
		reportMarketCBORWireStats(b, stats)
	})
	b.Run(name+"/decode", func(b *testing.B) {
		index := 0
		b.ReportAllocs()
		for b.Loop() {
			value := values[index]
			benchU256Sink, _ = value.codec.ParseCBOR(value.wire)
			index++
			if index == len(values) {
				index = 0
			}
		}
		reportMarketCBORWireStats(b, stats)
	})
}

func appendPreparedMarketCBORBar(dst []byte, bar preparedMarketCBORBar) []byte {
	dst = append(dst, 0x8e, 0x01)
	dst = appendCBORTextOracle(dst, bar.symbol)
	for _, price := range bar.prices {
		dst = bar.priceCodec.AppendCBOR(dst, price)
	}
	dst = bar.quantityCodec.AppendCBOR(dst, bar.volume)
	dst = appendCBORUint64(dst, bar.firstUpdateID)
	dst = appendCBORUint64(dst, bar.lastUpdateID)
	dst = appendCBORUint64(dst, bar.updateCount)
	dst = appendCBORUint64(dst, bar.barStartMS)
	dst = appendCBORUint64(dst, bar.barCloseMS)
	dst = appendCBORUint64(dst, bar.finalized)
	return appendCBORUint64(dst, uint64(bar.flags))
}

func decodePreparedMarketCBORBar(
	raw []byte,
	priceCodec Uint256Codec,
	quantityCodec Uint256Codec,
) (preparedMarketCBORBar, error) {
	if len(raw) == 0 || raw[0] != 0x8e {
		return preparedMarketCBORBar{}, boxedErrCBORSyntax
	}
	cursor := cborBarOracleCursor{raw: raw[1:]}
	recordKind, parseErr := cursor.readUint(1)
	if parseErr != "" || recordKind != 1 {
		return preparedMarketCBORBar{}, boxedErrCBORSyntax
	}
	symbol, parseErr := cursor.readText()
	if parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	bar := preparedMarketCBORBar{priceCodec: priceCodec, quantityCodec: quantityCodec, symbol: symbol}
	for i := range bar.prices {
		var err Error
		cursor.raw, err = priceCodec.ParseCBORFirstInto(cursor.raw, &bar.prices[i])
		if err != "" {
			return preparedMarketCBORBar{}, boxedError(err)
		}
	}
	var decodeErr Error
	cursor.raw, decodeErr = quantityCodec.ParseCBORFirstInto(cursor.raw, &bar.volume)
	if decodeErr != "" {
		return preparedMarketCBORBar{}, boxedError(decodeErr)
	}
	if bar.firstUpdateID, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	if bar.lastUpdateID, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	if bar.updateCount, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	if bar.barStartMS, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	if bar.barCloseMS, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	if bar.finalized, parseErr = cursor.readUint(^uint64(0)); parseErr != "" {
		return preparedMarketCBORBar{}, boxedError(parseErr)
	}
	flags, parseErr := cursor.readUint(uint64(^uint16(0)))
	if parseErr != "" || len(cursor.raw) != 0 {
		return preparedMarketCBORBar{}, boxedErrCBORSyntax
	}
	bar.flags = uint16(flags)
	return bar, nil
}

type marketCBORWireStats struct {
	minimum int
	p50     int
	p95     int
	maximum int
	total   int
	count   int
}

func marketCBORScalarWireStats(values []preparedMarketCBORScalar) marketCBORWireStats {
	sizes := make([]int, len(values))
	for i, value := range values {
		sizes[i] = len(value.wire)
	}
	return calculateMarketCBORWireStats(sizes)
}

func marketCBORBarWireStats(bars []preparedMarketCBORBar) marketCBORWireStats {
	sizes := make([]int, len(bars))
	buffer := make([]byte, 0, 192)
	for i, bar := range bars {
		sizes[i] = len(appendPreparedMarketCBORBar(buffer[:0], bar))
	}
	return calculateMarketCBORWireStats(sizes)
}

func calculateMarketCBORWireStats(sizes []int) marketCBORWireStats {
	sorted := append([]int(nil), sizes...)
	sort.Ints(sorted)
	total := 0
	for _, size := range sorted {
		total += size
	}
	return marketCBORWireStats{
		minimum: sorted[0],
		p50:     sorted[(len(sorted)-1)*50/100],
		p95:     sorted[(len(sorted)-1)*95/100],
		maximum: sorted[len(sorted)-1],
		total:   total,
		count:   len(sorted),
	}
}

func reportMarketCBORWireStats(b *testing.B, stats marketCBORWireStats) {
	b.ReportMetric(float64(stats.count), "samples")
	b.ReportMetric(float64(stats.minimum), "min-B/wire")
	b.ReportMetric(float64(stats.p50), "p50-B/wire")
	b.ReportMetric(float64(stats.p95), "p95-B/wire")
	b.ReportMetric(float64(stats.maximum), "max-B/wire")
	b.ReportMetric(float64(stats.total)/float64(stats.count), "mean-B/wire")
}

func compareUint256(left, right uint256.Int) int {
	for i := len(left) - 1; i >= 0; i-- {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	return 0
}
