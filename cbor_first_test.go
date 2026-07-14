package sailfish

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/holiman/uint256"
)

type cborBarOracle struct {
	_ struct{} `cbor:",toarray"`

	RecordKind        uint8
	Symbol            string
	Open              Decimal[PriceUint64[Fraction5], uint64]
	High              Decimal[PriceUint64[Fraction5], uint64]
	Low               Decimal[PriceUint64[Fraction5], uint64]
	Close             Decimal[PriceUint64[Fraction5], uint64]
	Volume            Decimal[AmountUint256[Fraction18], uint256.Int]
	FirstUpdateID     uint64
	LastUpdateID      uint64
	UpdateCount       uint64
	BarStartMS        uint64
	BarCloseMS        uint64
	FinalizedAppBlock uint64
	Flags             uint16
}

var (
	cborBarPriceCodec  = testCodec[PriceUint64[Fraction5]]()
	cborBarAmountCodec = testCodec[AmountUint256[Fraction18]]()
)

func TestCBORFirstDecodesManualPositionalFields(t *testing.T) {
	t.Parallel()

	open := cborBarPriceCodec.FromUnits(4_321_012_345)
	volume := cborBarAmountCodec.FromUnits(uint256.Int{0x5f6a_b7ce_3c7e_9a00, 6})

	wire := []byte{0x82}
	wire = cborBarPriceCodec.AppendCBOR(wire, open)
	wire = cborBarAmountCodec.AppendCBOR(wire, volume)

	decodedOpen, rest, err := cborBarPriceCodec.ParseCBORFirst(wire[1:])
	if err != nil {
		t.Fatal(err)
	}
	decodedVolume, rest, err := cborBarAmountCodec.ParseCBORFirst(rest)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 0 {
		t.Fatalf("remaining bytes = %x", rest)
	}
	if !decodedOpen.Equal(open) || !decodedVolume.Equal(volume) {
		t.Fatalf("decoded = %v, %v", decodedOpen.Units(), decodedVolume.Units())
	}

	if _, err = cborBarPriceCodec.ParseCBOR(wire[1:]); !errors.Is(err, ErrCBORSyntax) {
		t.Fatalf("exact ParseCBOR error = %v, want %v", err, ErrCBORSyntax)
	}
}

func TestCBORManualBarMatchesFxamackerToArray(t *testing.T) {
	t.Parallel()

	value := cborBarFixture()
	manual := appendManualCBORBarOracle(make([]byte, 0, 93), value)
	fxWire, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := fxWire.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(manual, oracle) {
		t.Fatalf("manual = %x\nfxamacker = %x", manual, oracle)
	}
	if got := len(manual); got != 93 {
		t.Fatalf("wire length = %d, want 93", got)
	}
	const golden = "8e0167535058555344431b00000001018d6a791b00000001022600f91b0000000100f4d3f91b0000000101d9b5b9c249065f6ab7ce3c7e9a001a000186a01a0001871f18801b0000018fcf6873001b0000018fcf695d601a00bc614e01"
	if got := hex.EncodeToString(manual); got != golden {
		t.Fatalf("wire = %s\nwant = %s", got, golden)
	}
}

func TestCBORFirstRejectsMalformedAndOverflowWithoutConsuming(t *testing.T) {
	t.Parallel()

	nativeCodec := testCodec[PriceUint8[Fraction1]]()
	for _, tt := range []struct {
		name string
		wire []byte
		err  error
	}{
		{name: "empty", wire: nil, err: ErrCBORSyntax},
		{name: "truncated", wire: []byte{0x19, 0x01}, err: ErrCBORSyntax},
		{name: "non-preferred", wire: []byte{0x18, 0x17, 0x01}, err: ErrCBORNonDeterministic},
		{name: "overflow", wire: []byte{0x19, 0x01, 0x00, 0x01}, err: ErrRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			decoded, rest, err := nativeCodec.ParseCBORFirst(tt.wire)
			if !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			if decoded.Units() != 0 || rest != nil {
				t.Fatalf("failure returned value=%d rest=%x", decoded.Units(), rest)
			}
		})
	}
}

func TestUint256CodecCBORFirstAndInto(t *testing.T) {
	t.Parallel()

	codec := testUint256Codec(18)
	want := uint256.Int{1, 2, 3, 4}
	var buffer [MaxCBORSize + 1]byte
	wire := codec.AppendCBOR(buffer[:0], want)
	wire = append(wire, 0x01)

	got, rest, err := codec.ParseCBORFirst(wire)
	if err != "" || got != want || !bytes.Equal(rest, []byte{0x01}) {
		t.Fatalf("ParseCBORFirst = %#v, %x, %v", got, rest, err)
	}

	var into uint256.Int
	rest, err = codec.ParseCBORFirstInto(wire, &into)
	if err != "" || into != want || !bytes.Equal(rest, []byte{0x01}) {
		t.Fatalf("ParseCBORFirstInto = %#v, %x, %v", into, rest, err)
	}
	before := into
	rest, err = codec.ParseCBORFirstInto([]byte{0xc2}, &into)
	if err != ErrCBORSyntax || rest != nil || into != before {
		t.Fatalf("failed ParseCBORFirstInto = %#v, %x, %v", into, rest, err)
	}
	rest, err = codec.ParseCBORFirstInto(wire, nil)
	if err != ErrNilDestination || rest != nil {
		t.Fatalf("nil ParseCBORFirstInto = %x, %v", rest, err)
	}
}

func cborBarFixture() cborBarOracle {
	return cborBarOracle{
		RecordKind:        1,
		Symbol:            "SPXUSDC",
		Open:              cborBarPriceCodec.FromUnits(4_321_012_345),
		High:              cborBarPriceCodec.FromUnits(4_331_012_345),
		Low:               cborBarPriceCodec.FromUnits(4_311_012_345),
		Close:             cborBarPriceCodec.FromUnits(4_326_012_345),
		Volume:            cborBarAmountCodec.FromUnits(uint256.Int{0x5f6a_b7ce_3c7e_9a00, 6}),
		FirstUpdateID:     100_000,
		LastUpdateID:      100_127,
		UpdateCount:       128,
		BarStartMS:        1_717_171_680_000,
		BarCloseMS:        1_717_171_740_000,
		FinalizedAppBlock: 12_345_678,
		Flags:             1,
	}
}

func appendManualCBORBarOracle(dst []byte, value cborBarOracle) []byte {
	dst = append(dst, 0x8e)
	dst = appendCBORUint64(dst, uint64(value.RecordKind))
	dst = appendCBORTextOracle(dst, value.Symbol)
	dst = cborBarPriceCodec.AppendCBOR(dst, value.Open)
	dst = cborBarPriceCodec.AppendCBOR(dst, value.High)
	dst = cborBarPriceCodec.AppendCBOR(dst, value.Low)
	dst = cborBarPriceCodec.AppendCBOR(dst, value.Close)
	dst = cborBarAmountCodec.AppendCBOR(dst, value.Volume)
	dst = appendCBORUint64(dst, value.FirstUpdateID)
	dst = appendCBORUint64(dst, value.LastUpdateID)
	dst = appendCBORUint64(dst, value.UpdateCount)
	dst = appendCBORUint64(dst, value.BarStartMS)
	dst = appendCBORUint64(dst, value.BarCloseMS)
	dst = appendCBORUint64(dst, value.FinalizedAppBlock)
	return appendCBORUint64(dst, uint64(value.Flags))
}

func appendCBORTextOracle(dst []byte, value string) []byte {
	if len(value) <= 23 {
		dst = append(dst, 0x60|byte(len(value)))
		return append(dst, value...)
	}
	dst = append(dst, 0x78, byte(len(value)))
	return append(dst, value...)
}

func decodeManualCBORBarOracle(raw []byte) (cborBarOracle, error) {
	if len(raw) == 0 || raw[0] != 0x8e {
		return cborBarOracle{}, boxedErrCBORSyntax
	}
	cursor := cborBarOracleCursor{raw: raw[1:]}
	recordKind, err := cursor.readUint(math.MaxUint8)
	if err != "" {
		return cborBarOracle{}, boxedError(err)
	}
	symbol, err := cursor.readText()
	if err != "" {
		return cborBarOracle{}, boxedError(err)
	}

	open, rest, decodeErr := cborBarPriceCodec.ParseCBORFirst(cursor.raw)
	if decodeErr != nil {
		return cborBarOracle{}, decodeErr
	}
	cursor.raw = rest
	high, rest, decodeErr := cborBarPriceCodec.ParseCBORFirst(cursor.raw)
	if decodeErr != nil {
		return cborBarOracle{}, decodeErr
	}
	cursor.raw = rest
	low, rest, decodeErr := cborBarPriceCodec.ParseCBORFirst(cursor.raw)
	if decodeErr != nil {
		return cborBarOracle{}, decodeErr
	}
	cursor.raw = rest
	closeValue, rest, decodeErr := cborBarPriceCodec.ParseCBORFirst(cursor.raw)
	if decodeErr != nil {
		return cborBarOracle{}, decodeErr
	}
	cursor.raw = rest
	volume, rest, decodeErr := cborBarAmountCodec.ParseCBORFirst(cursor.raw)
	if decodeErr != nil {
		return cborBarOracle{}, decodeErr
	}
	cursor.raw = rest

	firstUpdateID, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	lastUpdateID, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	updateCount, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	barStartMS, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	barCloseMS, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	finalizedAppBlock, parseErr := cursor.readUint(math.MaxUint64)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	flags, parseErr := cursor.readUint(math.MaxUint16)
	if parseErr != "" {
		return cborBarOracle{}, boxedError(parseErr)
	}
	if len(cursor.raw) != 0 {
		return cborBarOracle{}, boxedErrCBORSyntax
	}

	return cborBarOracle{
		RecordKind:        uint8(recordKind),
		Symbol:            symbol,
		Open:              open,
		High:              high,
		Low:               low,
		Close:             closeValue,
		Volume:            volume,
		FirstUpdateID:     firstUpdateID,
		LastUpdateID:      lastUpdateID,
		UpdateCount:       updateCount,
		BarStartMS:        barStartMS,
		BarCloseMS:        barCloseMS,
		FinalizedAppBlock: finalizedAppBlock,
		Flags:             uint16(flags),
	}, nil
}

type cborBarOracleCursor struct {
	raw []byte
}

func (c *cborBarOracleCursor) readUint(maxValue uint64) (uint64, Error) {
	value, consumed, err := parseCBORUint64First(c.raw, maxValue)
	if err != "" {
		return 0, err
	}
	c.raw = c.raw[consumed:]
	return value, ""
}

func (c *cborBarOracleCursor) readText() (string, Error) {
	if len(c.raw) == 0 || c.raw[0]>>5 != 3 {
		return "", ErrCBORSyntax
	}
	additional := c.raw[0] & 0x1f
	headerLen := 1
	textLen := int(additional)
	if additional == cborUnsignedAdditionalUint8 {
		if len(c.raw) < 2 {
			return "", ErrCBORSyntax
		}
		if c.raw[1] <= 23 {
			return "", ErrCBORNonDeterministic
		}
		headerLen = 2
		textLen = int(c.raw[1])
	} else if additional > 23 {
		return "", ErrCBORSyntax
	}
	consumed := headerLen + textLen
	if len(c.raw) < consumed {
		return "", ErrCBORSyntax
	}
	value := string(c.raw[headerLen:consumed])
	c.raw = c.raw[consumed:]
	return value, ""
}
