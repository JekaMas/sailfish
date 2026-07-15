package sailfish

import (
	"math/big"
	"math/bits"
	"testing"

	"github.com/holiman/uint256"
)

var benchmarkRatSink big.Rat
var benchmarkDenominatedSink Denominated[uint32, PriceInUint64Units[DecimalPlaces5], uint64]
var benchmarkAssetDenominatedSink Denominated[testAsset, PriceInUint64Units[DecimalPlaces5], uint64]
var benchmarkScaleInputs = [...]uint64{120, 3, 1_000, 100}

func BenchmarkBigRatCeilings(b *testing.B) {
	numerator := new(big.Int).SetUint64(12_331_232)
	denominator := new(big.Int).SetUint64(100_000)
	wide := uint256.Int{1, 2, 3, 4}
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	rational64 := big.NewRat(3_082_808, 25_000)
	rational256 := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 200))

	b.Run("set_frac64_reused", func(b *testing.B) {
		var destination big.Rat
		destination.SetFrac64(12_331_232, 100_000)
		b.ResetTimer()
		for b.Loop() {
			destination.SetFrac64(12_331_232, 100_000)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("set_frac_big_reused", func(b *testing.B) {
		var destination big.Rat
		destination.SetFrac(numerator, denominator)
		b.ResetTimer()
		for b.Loop() {
			destination.SetFrac(numerator, denominator)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("set_uint64_reused", func(b *testing.B) {
		var destination big.Rat
		destination.SetUint64(123)
		b.ResetTimer()
		for b.Loop() {
			destination.SetUint64(123)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("set_int_reused", func(b *testing.B) {
		integer := new(big.Int).Lsh(big.NewInt(1), 200)
		var destination big.Rat
		destination.SetInt(integer)
		b.ResetTimer()
		for b.Loop() {
			destination.SetInt(integer)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("read_num_denom_u256", func(b *testing.B) {
		source := new(big.Rat).SetFrac(numerator, denominator)
		for b.Loop() {
			var num, den uint256.Int
			if num.SetFromBig(source.Num()) || den.SetFromBig(source.Denom()) {
				b.Fatal("unexpected overflow")
			}
			benchU256Sink = num
		}
	})
	b.Run("holiman_mul_overflow", func(b *testing.B) {
		factor := uint256.Int{100_000}
		for b.Loop() {
			var result uint256.Int
			if _, overflow := result.MulOverflow(&wide, &factor); overflow {
				b.Fatal("unexpected overflow")
			}
			benchU256Sink = result
		}
	})
	b.Run("scale_lookup/checked_uint64", func(b *testing.B) {
		for b.Loop() {
			scale, err := checkedFractionalDecimalPlaces[PriceInUint64Units[DecimalPlaces5], uint64]()
			benchIntSink = scale + int(len(err))
		}
	})
	b.Run("scale_lookup/cached_uint64", func(b *testing.B) {
		for b.Loop() {
			benchIntSink = codec64.fractionalDecimalPlaces()
		}
	})
	b.Run("from_rat_native/direct_result", func(b *testing.B) {
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = codec64.fromBigRatNative(rational64, 5)
		}
	})
	b.Run("from_rat_native/via_u256", func(b *testing.B) {
		for b.Loop() {
			numerator := rational64.Num().Uint64()
			denominator := rational64.Denom().Uint64()
			units := numerator * (powersOf10Uint64[5] / denominator)
			allocationPriceSink, allocationErrorSink = codec64.FromU256(uint256.Int{units})
		}
	})
	b.Run("from_rat_wide/checked_scale", func(b *testing.B) {
		for b.Loop() {
			scale, _ := checkedFractionalDecimalPlaces[AmountInUint256Units[DecimalPlaces18], uint256.Int]()
			integerConversionWideSink, allocationErrorSink = codec256.fromBigRatUint256(rational256, scale)
		}
	})
	b.Run("from_rat_wide/cached_scale", func(b *testing.B) {
		for b.Loop() {
			integerConversionWideSink, allocationErrorSink = codec256.fromBigRatUint256(
				rational256,
				codec256.fractionalDecimalPlaces(),
			)
		}
	})
	b.Run("to_rat_native/public_fractional", func(b *testing.B) {
		value := codec64.FromUnits(12_331_232)
		var destination big.Rat
		var workspace BigRatWorkspace
		for b.Loop() {
			allocationErrorSink = value.ToBigRat(&destination, &workspace)
		}
	})
	b.Run("to_rat_native/baseline_fractional", func(b *testing.B) {
		value := codec64.FromUnits(12_331_232)
		var destination big.Rat
		var workspace BigRatWorkspace
		for b.Loop() {
			allocationErrorSink = benchmarkToBigRatWithoutIntegralFastPath(value, &destination, &workspace)
		}
	})
	b.Run("to_rat_wide/public_fractional", func(b *testing.B) {
		value := codec256.FromUnits(wide)
		var destination big.Rat
		var workspace BigRatWorkspace
		for b.Loop() {
			allocationErrorSink = value.ToBigRat(&destination, &workspace)
		}
	})
	b.Run("to_rat_wide/baseline_fractional", func(b *testing.B) {
		value := codec256.FromUnits(wide)
		var destination big.Rat
		var workspace BigRatWorkspace
		for b.Loop() {
			allocationErrorSink = benchmarkToBigRatWithoutIntegralFastPath(value, &destination, &workspace)
		}
	})
}

func benchmarkToBigRatWithoutIntegralFastPath[V FixedDecimalFormat[U], U Unit](
	value FixedDecimal[V, U],
	destination *big.Rat,
	workspace *BigRatWorkspace,
) error {
	scale, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return boxedError(err)
	}
	if err := value.ToBigInt(&workspace.numerator); err != nil {
		return err
	}
	if scale <= maxUint64Scale {
		workspace.denominator.SetUint64(powersOf10Uint64[scale])
	} else {
		denominator := &workspace.denominator
		powersOf10Uint256[scale].IntoBig(&denominator)
	}
	destination.SetFrac(&workspace.numerator, &workspace.denominator)
	return nil
}

func BenchmarkCrossScaleCeilings(b *testing.B) {
	left, right := benchmarkScaleInputs[0], benchmarkScaleInputs[1]
	leftFactor, rightFactor := benchmarkScaleInputs[2], benchmarkScaleInputs[3]
	b.Run("uint64_multiply_checked", func(b *testing.B) {
		for b.Loop() {
			hi, lo := bits.Mul64(left, leftFactor)
			benchUint64Sink = lo | hi
		}
	})
	b.Run("uint64_mixed_scale_add_checked", func(b *testing.B) {
		for b.Loop() {
			leftHi, scaledLeft := bits.Mul64(left, leftFactor)
			rightHi, scaledRight := bits.Mul64(right, rightFactor)
			sum, carry := bits.Add64(scaledLeft, scaledRight, 0)
			benchUint64Sink = sum | leftHi | rightHi | carry
		}
	})
}

func BenchmarkRationalAndCrossScaleOperations(b *testing.B) {
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	rational64 := big.NewRat(3_082_808, 25_000)
	rational256 := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 200))
	value64 := codec64.FromUnits(12_331_232)
	value256 := codec256.FromUnits(uint256.Int{1, 2, 3, 4})
	left := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]().FromUnits(120)
	right := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces3]]().FromUnits(3)

	b.Run("from_rat/uint64", func(b *testing.B) {
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = codec64.FromBigRat(rational64)
		}
	})
	b.Run("from_rat/uint256", func(b *testing.B) {
		for b.Loop() {
			integerConversionWideSink, allocationErrorSink = codec256.FromBigRat(rational256)
		}
	})
	b.Run("to_rat_reused/uint64", func(b *testing.B) {
		var destination big.Rat
		var workspace BigRatWorkspace
		_ = value64.ToBigRat(&destination, &workspace)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value64.ToBigRat(&destination, &workspace)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("to_rat_reused/uint256", func(b *testing.B) {
		var destination big.Rat
		var workspace BigRatWorkspace
		_ = value256.ToBigRat(&destination, &workspace)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value256.ToBigRat(&destination, &workspace)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("to_rat_reused/integer_uint64", func(b *testing.B) {
		value := codec64.FromUnits(12_300_000)
		var destination big.Rat
		var workspace BigRatWorkspace
		_ = value.ToBigRat(&destination, &workspace)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value.ToBigRat(&destination, &workspace)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("to_rat_reused/integer_uint256", func(b *testing.B) {
		value := codec256.FromUnits(uint256.Int{1_000_000_000_000_000_000})
		var destination big.Rat
		var workspace BigRatWorkspace
		_ = value.ToBigRat(&destination, &workspace)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value.ToBigRat(&destination, &workspace)
		}
		benchmarkRatSink.Set(&destination)
	})
	b.Run("rescale/uint64_2_to_5", func(b *testing.B) {
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = Rescale[PriceInUint64Units[DecimalPlaces5]](left)
		}
	})
	b.Run("add_as/uint64_mixed_scales", func(b *testing.B) {
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = AddAs[PriceInUint64Units[DecimalPlaces5]](left, right)
		}
	})
	b.Run("denominated/add_same_scale", func(b *testing.B) {
		left := NewDenominated(uint32(7), value64)
		right := NewDenominated(uint32(7), codec64.FromUnits(1))
		for b.Loop() {
			benchmarkDenominatedSink, allocationErrorSink = left.Add(right)
		}
	})
	b.Run("denominated/add_same_scale_asset_identity", func(b *testing.B) {
		asset := testAsset{Chain: 1, Token: "USDC"}
		left := NewDenominated(asset, value64)
		right := NewDenominated(asset, codec64.FromUnits(1))
		for b.Loop() {
			benchmarkAssetDenominatedSink, allocationErrorSink = left.Add(right)
		}
	})
	b.Run("denominated/add_as_mixed_scales", func(b *testing.B) {
		left := NewDenominated(uint32(7), left)
		right := NewDenominated(uint32(7), right)
		for b.Loop() {
			benchmarkDenominatedSink, allocationErrorSink = AddDenominatedAs[PriceInUint64Units[DecimalPlaces5]](left, right)
		}
	})
}
