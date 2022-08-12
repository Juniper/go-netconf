package netconf

import "encoding/xml"

type GetConfig struct {
	XMLName xml.Name   `xml:"get-config"`
	Source  StringElem `xml:"source"`
	Filter  *Filter    `xml:"filter,omitempty"`
}

// this really should be handled by session.Close and so is not-exported.
type closeSession struct {
	XMLName xml.Name `xml:"close-session"`
}

type Filter struct {
	XMLName xml.Name `xml:"filter"`
	Type    xml.Attr `xml:"type"`
}
