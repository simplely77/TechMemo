package dto

import "time"

type CreateNoteReq struct {
	Title      string  `json:"title"`
	ContentMD  string  `josn:"content_md"`
	CategoryID int64   `json:"category_id"`
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
}
