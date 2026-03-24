package dao

type GetNotesParams struct {
	UserID     int64
	CategoryID int64
	TagIDs     []int64
	Keyword    string
	NoteType   string
	Sort       string
	Limit      int64
	Offset     int64
}

type UpdateNoteParams struct {
	Title      *string
	ContentMD  *string
	CategoryID *int64
	NoteType   *string
}

type CreateAILogParams struct {
	SourceNoteID int64
	TaskID       string
	TargetType   string
	TargetID     int64
	ProcessType  string
	ModelName    string
	Status       string
}

type GetKnowledgePointsParams struct {
	UserID        int64
	SourceNoteID  int64
	Keyword       string
	MinImportance float64
	Page          int64
	PageSize      int64
}

type UpdateKnowledgePointParams struct {
	Name            string
	Description     string
	ImportanceScore float64
}

// CreateSessionParams 创建会话的参数
type CreateSessionParams struct {
	UserID int64
	Title  string
}

// CreateMessageParams 创建消息的参数
type CreateMessageParams struct {
	SessionID  int64
	UserID     int64
	Role       string
	Content    string
	TokenCount int32
}
