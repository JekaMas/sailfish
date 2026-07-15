//go:build !amd64 && !arm64

package sailfish

// loadEightBytes is the portable, endian-independent SWAR load. Callers
// establish that bytes [offset:offset+8] exist before calling.
func loadEightBytes[S decimalInput](input S, offset int) uint64 {
	return uint64(input[offset]) |
		uint64(input[offset+1])<<8 |
		uint64(input[offset+2])<<16 |
		uint64(input[offset+3])<<24 |
		uint64(input[offset+4])<<32 |
		uint64(input[offset+5])<<40 |
		uint64(input[offset+6])<<48 |
		uint64(input[offset+7])<<56
}
