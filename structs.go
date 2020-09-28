package main

// FileData represents a file stored in hexFS.
type FileData struct {
	ID string `json:"id,omitempty," bson:"id,omitempty"`
	Ext string `json:"ext,omitempty" bson:"ext,omitempty"`
	SHA256 string `json:"sha256,omitempty" bson:"sha256,omitempty"`
	UploadedTimestamp string `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	IP string `json:"ip,omitempty" bson:"ip,omitempty"`
	Size int64 `json:"size,omitempty" bson:"size,omitempty"`
}