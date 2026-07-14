// Package sailfish provides fast, unsigned, fixed-scale decimal values for
// trading and financial protocols.
//
// A value is stored as one scaled integer:
//
//	value = units / 10^scale
//
// The package supports uint64 and uint256.Int units. Its hot parse, append,
// compare, and arithmetic paths are allocation-free when caller-owned output
// buffers have enough capacity.
package sailfish
