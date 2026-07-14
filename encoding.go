package sailfish

import (
	"bytes"

	"github.com/goccy/go-json"
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
	out := make([]byte, 0, d.Len()+2)
	return d.AppendJSON(out), nil
}

// UnmarshalJSON uses an allocation-free fast path for ordinary unescaped JSON
// strings and go-json only for escaped strings.
func (d *Decimal[V, U]) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		text := data[1 : len(data)-1]
		if bytes.IndexByte(text, '\\') < 0 {
			parsed, err := NewBytes[V](text)
			if err != nil {
				return err
			}
			*d = parsed
			return nil
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
