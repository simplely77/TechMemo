package dto

import (
	"time"
)

type CreateNoteReq struct {
	Title      string  `json:"title" binding:"required"`
	ContentMD  string  `josn:"content_md"`
	CategoryID int64   `json:"category_id" binding:"required"`
	TagIDs     []int64 `json:"tag_ids"`
}

type CreateNoteResp struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	ContentMD  string    `json:"content_md"`
	NoteType   string    `json:"note_type"`
	CategoryID int64     `json:"category_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type GetNotesReq struct {
	CategoryID int64   `form:"category_id"`
	TagIDs     []int64 `form:"tag_ids" binding:"omitempty,dive,number"`
	Keyword    string  `form:"keyword"`
	NoteType   string  `form:"note_type"`
	Page       int64   `form:"page,default=1"`
	PageSize   int64   `form:"page_size,default=20"`
	Sort       string  `form:"sort,default=created_at_desc"`
}

type NoteCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type NoteTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type NoteAbstract struct {
	ID        int64        `json:"id"`
	Title     string       `json:"title"`
	NoteType  string       `json:"note_type"`
	Category  NoteCategory `json:"category"`
	Tags      []NoteTag    `json:"tags"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type GetNotesResp struct {
	Notes    []*NoteAbstract `json:"notes"`
	Total    int64           `json:"total"`
	Page     int64           `json:"page"`
	PageSize int64           `json:"page_size"`
}

type GetNoteResp struct {
	ID        int64        `json:"id"`
	Title     string       `json:"title"`
	ContentMD string       `json:"content_md"`
	NoteType  string       `json:"note_type"`
	Category  NoteCategory `json:"category"`
	Tags      []NoteTag    `json:"tags"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type UpdateNoteReq struct {
	Title      *string `json:"title"`
	ContentMD  *string `json:"content_md"`
	CategoryID *int64  `json:"category_id"`
}

type UpdateNoteTagsReq struct {
	TagIDs []int64 `json:"tag_ids"`
}
type Version struct {
	ID        int64     `json:"id"`
	NoteID    int64     `json:"note_id"`
	ContentMD string    `json:"content_md"`
	CreatedAt time.Time `json:"created_at"`
}

type GetNoteVersionsResp struct {
	Versions []Version `json:"versions"`
}
