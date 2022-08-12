package netconf

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type StringElem string

func (s StringElem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if s == "" {
		return fmt.Errorf("string elements cannot be empty")
	}

	escaped, err := escapeXML(string(s))
	if err != nil {
		return fmt.Errorf("invalid string element: %w", err)
	}

	v := struct {
		Elem string `xml:",innerxml"`
	}{Elem: "<" + escaped + "/>"}
	return e.EncodeElement(&v, start)
}

type SentinalBool bool

func (b SentinalBool) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if !b {
		return nil
	}
	return e.EncodeElement(b, start)
}

func (b *SentinalBool) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := &struct{}{}
	if err := d.DecodeElement(v, &start); err != nil {
		return err
	}
	*b = v != nil
	return nil
}

func escapeXML(input string) (string, error) {
	buf := &strings.Builder{}
	if err := xml.EscapeText(buf, []byte(input)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
