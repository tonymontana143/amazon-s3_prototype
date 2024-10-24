package models

import "encoding/xml"

type Bucket struct {
	XMLName          xml.Name `xml:"CreateBucketConfiguration"`
	Name             string   `xml:"Bucket>Name"`
	CreationTime     string   `xml:"Bucket>CreationTime"`
	LastModifiedTime string   `xml:"Bucket>LastModifiedTime"`
	Status           string   `xml:"Bucket>Status"`
}
