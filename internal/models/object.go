package models

import "encoding/xml"

type Object struct {
	XMLName          xml.Name `xml:"PutObjectConfiguration"`
	ObjectKey        string   `xml:"Object>ObjectKey"`
	Size             int64    `xml:"Object>Size"`
	ContentType      string   `xml:"Object>ContentType"`
	LastModifiedTime string   `xml:"Object>LastModifiedTime"`
}
