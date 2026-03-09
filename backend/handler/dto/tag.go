package dto

type Tag struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	UserID int64  `json:"user_id"`
}

type GetTagsReq struct {
	Keyword  string `form:"keyword"`
	Page     int64  `form:"page,default=1"`
	PageSize int64  `form:"page_size,default=20"`
}

type GetTagsResp struct {
	Tags     []Tag `json:"tags"`
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type CreateTagReq struct {
	Name string `json:"name"`
}

type UpdateTagReq struct {
	Name string `json:"name"`
}
