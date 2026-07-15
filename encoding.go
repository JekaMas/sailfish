package sailfish

import (
	"bytes"

	"github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

// AppendText implements the append-style text encoding contract available in
// current Go versions without requiring a newly owned result slice.
func (d Decimal[V, U]) AppendText(dst []byte) ([]byte, error) {
	return d.AppendTo(dst), nil
}

func (d Decimal[V, U]) MarshalText() ([]byte, error) {
	out := make([]byte, 0, d.Len())
	return d.AppendTo(out), nil
}

func (d *Decimal[V, U]) UnmarshalText(text []byte) error {
	parsed, err := NewBytes[V](text)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Decimal[V, U]) MarshalJSON() ([]byte, error) {
	capacity := 0
	if d.representation == "" {
		var units U
		if _, wide := any(units).(uint256.Int); wide {
			// Computing the exact text length of a wide value performs the same
			// decimal split as formatting. Reserve the bounded maximum so raw
			// uint256 values are split only once while retaining one owned
			// result allocation.
			capacity = maxUint256TextLen + 2
		}
	}
	if capacity == 0 {
		capacity = d.Len() + 2
	}
	out := make([]byte, 0, capacity)
	return d.AppendJSON(out), nil
}

// UnmarshalJSON parses ordinary quoted decimals directly without a separate
// escape scan. go-json handles escaped strings and non-string JSON syntax.
func (d *Decimal[V, U]) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		text := data[1 : len(data)-1]
		parsed, err := NewBytes[V](text)
		if err == nil {
			*d = parsed
			return nil
		}
		if bytes.IndexByte(text, '\\') < 0 {
			return err
		}
	}

	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	parsed, err := New[V](text)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
