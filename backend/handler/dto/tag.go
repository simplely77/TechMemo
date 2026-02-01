package dto

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type GetTagsResp struct {
	Tags     []Tag `json:"tags"`
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type CreateTag struct {
	Name string `json:"name"`
}
