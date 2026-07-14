// Package sailfish provides fast, unsigned, fixed-scale decimal values for
// trading and financial protocols.
//
// A value is stored as one scaled integer:
//
//	value = units / 10^scale
//
// The package supports uint8, uint16, uint32, uint64, and uint256.Int units.
// Fractional scale and backend capacity are selected independently. Its hot
// parse, text/CBOR append, strict CBOR decode, compare, and arithmetic paths
// are allocation-free when caller-owned output buffers have enough capacity.
package sailfish
