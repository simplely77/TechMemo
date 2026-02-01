package dto

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type GetCategoriesResp struct {
	Categories []Category `json:"categories"`
}

type CreateCategoryReq struct {
	Name string `json:"name"`
}

type CreateCategoryResp struct {
	Category Category
}

type UpdateCategoryReq struct {
	Name string `json:"name"`
}
