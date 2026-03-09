package dto

type AIProcessResp struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type GetNoteAIStatusResp struct {
	NoteID   int64             `json:"note_id"`
	Status   string            `json:"status"`
	Progress map[string]string `json:"progress,omitempty"`
}
