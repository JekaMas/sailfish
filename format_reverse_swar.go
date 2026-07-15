package sailfish

import "encoding/binary"

const decimal1e8 = uint64(100_000_000)

// packedASCII8 converts value to exactly eight zero-padded ASCII digits in a
// uint64 whose little-endian memory representation is the output text.
//
// The arithmetic is reverse SWAR: two independent four-digit lanes are
// divided by 100 and then four two-digit lanes by 10 using exact reciprocal
// multiplication. This replaces the serial /100 chain and digit-pair table
// loads for complete eight-digit blocks. value must be below 1e8; all callers
// establish that through decimal block decomposition.
func packedASCII8(value uint32) uint64 {
	high := value / 10_000
	low := value - high*10_000
	lanes := uint64(high) | uint64(low)<<32

	quotient100 := ((lanes * 10_486) >> 20) & (uint64(0x7f)<<32 | 0x7f)
	remainder100 := lanes - quotient100*100
	pairs := remainder100<<16 + quotient100

	quotient10 := (pairs * 103) >> 10
	quotient10 &= uint64(0x0f)<<48 | uint64(0x0f)<<32 | uint64(0x0f)<<16 | 0x0f
	digits := quotient10 + (pairs-quotient10*10)<<8
	return digits + asciiZero8
}

// putLittleEndianExact writes the low 1-8 bytes of value without overstore.
// Full eight-byte stores are fastest, but exact-width tails preserve AppendTo's
// caller-owned-buffer contract; measured overstore was therefore rejected.
func putLittleEndianExact(dst []byte, value uint64) {
	switch len(dst) {
	case 1:
		dst[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(dst, uint16(value))
	case 3:
		binary.LittleEndian.PutUint16(dst, uint16(value))
		dst[2] = byte(value >> 16)
	case 4:
		binary.LittleEndian.PutUint32(dst, uint32(value))
	case 5:
		binary.LittleEndian.PutUint32(dst, uint32(value))
		dst[4] = byte(value >> 32)
	case 6:
		binary.LittleEndian.PutUint32(dst, uint32(value))
		binary.LittleEndian.PutUint16(dst[4:], uint16(value>>32))
	case 7:
		binary.LittleEndian.PutUint32(dst, uint32(value))
		binary.LittleEndian.PutUint16(dst[4:], uint16(value>>32))
		dst[6] = byte(value >> 48)
	case 8:
		binary.LittleEndian.PutUint64(dst, value)
	}
}

func putPackedWidth(dst []byte, value uint32) {
	packed := packedASCII8(value) >> ((8 - len(dst)) * 8)
	putLittleEndianExact(dst, packed)
}

// putPackedDigitsWithPoint converts a 1-7 digit scaled integer once and
// inserts the point in-register. The exact-width store exposes only the
// logical bytes and leaves caller-owned capacity untouched.
func putPackedDigitsWithPoint(dst []byte, value uint32, digits, integerDigits int) {
	packed := packedASCII8(value) >> ((8 - digits) * 8)
	lowMask := uint64(1)<<(integerDigits*8) - 1
	withPoint := packed&lowMask |
		uint64('.')<<(integerDigits*8) |
		packed&^lowMask<<8
	putLittleEndianExact(dst, withPoint)
}

// fillPacked64 decomposes a 14-20 digit uint64 into at most three 1e8 blocks.
// Lower blocks become independent full-width stores; only the leading block
// uses an exact-width tail. Benchmarks keep the pair-table formatter below the
// 14-digit crossover.
func fillPacked64(dst []byte, value uint64) {
	// Selected production widths are 14-20. One upfront proof lets the compiler
	// eliminate the dynamic lower-bound checks on both 1e8 block slices.
	_ = dst[13]
	width := len(dst)
	if width <= 16 {
		high := value / decimal1e8
		low := uint32(value - high*decimal1e8)
		putPackedWidth(dst[:width-8], uint32(high))
		binary.LittleEndian.PutUint64(dst[width-8:], packedASCII8(low))
		return
	}

	quotient := value / decimal1e8
	low := uint32(value - quotient*decimal1e8)
	high := quotient / decimal1e8
	middle := uint32(quotient - high*decimal1e8)
	putPackedWidth(dst[:width-16], uint32(high))
	binary.LittleEndian.PutUint64(dst[width-16:], packedASCII8(middle))
	binary.LittleEndian.PutUint64(dst[width-8:], packedASCII8(low))
}

// putPacked8WithPoint inserts a decimal point into one eight-digit packed
// word. The first eight output bytes are assembled in a register and stored at
// once; the displaced final digit is byte nine. integerDigits must be 1-7.
func putPacked8WithPoint(dst []byte, value uint32, integerDigits int) {
	// The output is always nine bytes. Proving the final byte once also removes
	// the independent check that binary.PutUint64 would otherwise retain.
	_ = dst[8]
	packed := packedASCII8(value)
	lowMask := uint64(1)<<(integerDigits*8) - 1
	withPoint := packed&lowMask |
		uint64('.')<<(integerDigits*8) |
		packed&^lowMask<<8
	binary.LittleEndian.PutUint64(dst, withPoint)
	dst[8] = byte(packed >> 56)
}
