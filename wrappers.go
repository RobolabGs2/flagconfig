package flagconfig

import (
	"encoding"
)

type encodingBinary interface {
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler
}

type binaryWrapper struct {
	binary encodingBinary
}

func (w binaryWrapper) String() string {
	if w.binary == nil {
		return ""
	}
	data, err := w.binary.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (w binaryWrapper) Set(s string) error {
	return w.binary.UnmarshalBinary([]byte(s))
}

type encodingText interface {
	encoding.TextUnmarshaler
	encoding.TextMarshaler
}

type textWrapper struct {
	text encodingText
}

func (w textWrapper) String() string {
	if w.text == nil {
		return ""
	}
	data, err := w.text.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (w textWrapper) Set(s string) error {
	return w.text.UnmarshalText([]byte(s))
}
