package sailfish

const (
	asciiZero8 = uint64(0x3030303030303030)
	ascii46x8  = uint64(0x4646464646464646)
	asciiHigh8 = uint64(0x8080808080808080)
)

// parseUint64Chunk parses at most 19 decimal digits with the scalar pairwise
// kernel used by native values and short wide-value prefixes.
func parseUint64Chunk[S decimalInput](s S, begin, end int) (uint64, Error) {
	var value uint64
	if (end-begin)&1 != 0 {
		digit := s[begin] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		value = uint64(digit)
		begin++
	}
	for ; begin < end; begin += 2 {
		a := s[begin] - '0'
		b := s[begin+1] - '0'
		if a > 9 || b > 9 {
			return 0, ErrSyntax
		}
		value = value*100 + uint64(a)*10 + uint64(b)
	}
	return value, ""
}

// parseUint64DenseChunk parses 8 through 19 decimal digits. Wide decimal
// parsing calls it only for dense chunks, keeping SWAR dispatch out of the
// latency-sensitive short native path.
func parseUint64DenseChunk[S decimalInput](s S, begin, end int) (uint64, Error) {
	remaining := end - begin
	if remaining < 8 {
		return parseUint64Chunk(s, begin, end)
	}

	// Reduce the first eight digits and, for 16-19 digit chunks, a second
	// independent block. Keeping these blocks explicit removes the loop branch
	// from the common 8/16/19-digit paths and lets the CPU overlap both SWAR
	// reductions before the single base-1e8 combination.
	first, ok := parseEightDigits(loadEightBytes(s, begin))
	if !ok {
		return 0, ErrSyntax
	}
	value := uint64(first)
	begin += 8
	if remaining >= 16 {
		second, valid := parseEightDigits(loadEightBytes(s, begin))
		if !valid {
			return 0, ErrSyntax
		}
		value = value*100_000_000 + uint64(second)
		begin += 8
	}
	if (end-begin)&1 != 0 {
		digit := s[begin] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		value = value*10 + uint64(digit)
		begin++
	}
	for ; begin < end; begin += 2 {
		a := s[begin] - '0'
		b := s[begin+1] - '0'
		if a > 9 || b > 9 {
			return 0, ErrSyntax
		}
		value = value*100 + uint64(a)*10 + uint64(b)
	}
	return value, ""
}

// parseEightDigits validates and reduces eight ASCII digits in one SWAR word.
// Three mask/multiply stages form pairs, four-digit groups, and the final
// uint32. This is the lowest-latency measured kernel for one short token on the
// supported arm64 host; it also avoids an assembly call and CPU-feature branch.
func parseEightDigits(raw uint64) (uint32, bool) {
	if ((raw+ascii46x8)|(raw-asciiZero8))&asciiHigh8 != 0 {
		return 0, false
	}
	raw = (raw - asciiZero8) * 2561 >> 8
	raw = (raw & 0x00ff00ff00ff00ff) * 6553601 >> 16
	raw = (raw & 0x0000ffff0000ffff) * 42949672960001 >> 32
	return uint32(raw), true
}
