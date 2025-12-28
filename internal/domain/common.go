package domain

// Pagination handles pagination data.
type Pagination struct {
	Limit  int64 `json:"limit" default:"10"`
	Offset int64 `json:"offset" default:"0"`
	Total  int64 `json:"total"`
}

// Lang handles multi-language strings.
type Lang struct {
	Uz string `json:"uz"`
	Ru string `json:"ru"`
	En string `json:"en"`
}

// File handles file metadata.
type File struct {
	Name string `json:"name"`
	Link string `json:"link"`
}
