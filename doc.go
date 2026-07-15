// Package sailfish provides fast, unsigned, fixed-decimal values for
// trading and financial protocols.
//
// A value is stored as one scaled integer:
//
//	value = units / 10^fractionalDecimalPlaces
//
// The package supports uint8, uint16, uint32, uint64, and uint256.Int units.
// Types make semantic kind, unit representation, and fractional decimal
// places explicit. For example, PriceInUint64Units[DecimalPlaces5] represents
// a price as uint64 units with exactly five digits after the decimal point.
// Its hot parse, text/CBOR append, strict CBOR decode, compare, and arithmetic
// paths are allocation-free when caller-owned output buffers have capacity.
package sailfish
