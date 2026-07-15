//go:build amd64 || arm64

package sailfish

import "unsafe"

// loadEightBytes performs one unaligned little-endian load for the validated
// SWAR parser on amd64 and arm64. Callers establish that bytes [offset:offset+8]
// exist before calling. The pointer is read-only, never retained, and does not
// create a mutable view of string storage. Build-time architecture selection
// avoids a feature branch in the nanosecond-scale parser.
func loadEightBytes[S decimalInput](input S, offset int) uint64 {
	switch value := any(input).(type) {
	case string:
		pointer := unsafe.Add(unsafe.Pointer(unsafe.StringData(value)), offset)
		return *(*uint64)(pointer)
	case []byte:
		return *(*uint64)(unsafe.Pointer(&value[offset]))
	default:
		// decimalInput is a closed exact-type set. This arm is unreachable, but
		// Go requires a return after a runtime type switch in generic code.
		return 0
	}
}
