package domain

// SortOrder represents the direction of sorting.
type SortOrder string

const (
	SortOrderASC  SortOrder = "ASC"
	SortOrderDESC SortOrder = "DESC"
)

// IsValid returns true if the SortOrder is ASC or DESC.
func (s SortOrder) IsValid() bool {
	return s == SortOrderASC || s == SortOrderDESC
}

// Pagination handles pagination data.
type Pagination struct {
	Limit     int64  `default:"10"      json:"limit"      binding:"min=1,max=1000"`
	Offset    int64  `default:"0"       json:"offset"     binding:"min=0"`
	Total     int64  `json:"total"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "ASC" or "DESC"
}

// Getters and Setters for Pagination
func (p *Pagination) GetLimit() int64        { return p.Limit }
func (p *Pagination) SetLimit(limit int64)   { p.Limit = limit }
func (p *Pagination) GetOffset() int64       { return p.Offset }
func (p *Pagination) SetOffset(offset int64) { p.Offset = offset }
func (p *Pagination) GetTotal() int64        { return p.Total }
func (p *Pagination) SetTotal(total int64)   { p.Total = total }

// Lang handles multi-language strings.
type Lang struct {
	Uz string `json:"uz"`
	Ru string `json:"ru"`
	En string `json:"en"`
}

// Getters and Setters for Lang
func (l *Lang) GetUz() string   { return l.Uz }
func (l *Lang) SetUz(uz string) { l.Uz = uz }
func (l *Lang) GetRu() string   { return l.Ru }
func (l *Lang) SetRu(ru string) { l.Ru = ru }
func (l *Lang) GetEn() string   { return l.En }
func (l *Lang) SetEn(en string) { l.En = en }

// File handles file metadata.
type File struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// Getters and Setters for File
func (f *File) GetName() string     { return f.Name }
func (f *File) SetName(name string) { f.Name = name }
func (f *File) GetLink() string     { return f.Link }
func (f *File) SetLink(link string) { f.Link = link }
