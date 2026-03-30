package service

import (
	"context"
	stderrors "errors"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"

	"gorm.io/gorm"
)

type CategoryService struct {
	categoryDao *dao.CategoryDao
}

func (c *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	err := c.categoryDao.DeleteCategory(ctx, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.CategoryNotFound
		}
		return errors.InternalErr
	}
	return nil
}

func (c *CategoryService) UpdateCategory(ctx context.Context, userID int64, id int64, name string) (*dto.Category, error) {
	exists, err := c.categoryDao.CheckCategoryExists(ctx, userID, name)
	if err != nil {
		return nil, errors.InternalErr
	}

	if exists {
		return nil, errors.CategoryExists
	}
	err = c.categoryDao.UpdateCategory(ctx, id, name)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.CategoryNotFound
		}
		return nil, errors.InternalErr
	}

	// 返回更新后的分类
	category, err := c.categoryDao.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr
	}

	return &dto.Category{
		ID:     category.ID,
		Name:   category.Name,
		UserID: category.UserID,
	}, nil
}

func (c *CategoryService) CreateCategory(ctx context.Context, userID int64, name string) (*dto.CreateCategoryResp, error) {
	exists, err := c.categoryDao.CheckCategoryExists(ctx, userID, name)
	if err != nil {
		return nil, errors.InternalErr
	}

	if exists {
		return nil, errors.CategoryExists
	}

	category := &model.Category{
		Name:   name,
		UserID: userID,
	}

	err = c.categoryDao.CreateCategory(ctx, category)
	if err != nil {
		return nil, errors.InternalErr
	}

	dtoCategory := dto.Category{
		ID:     category.ID,
		Name:   category.Name,
		UserID: category.UserID,
	}
	return &dto.CreateCategoryResp{Category: dtoCategory}, nil
}

func (c *CategoryService) GetCategorys(ctx context.Context, userID int64) (*dto.GetCategoriesResp, error) {
	categories, err := c.categoryDao.GetCategoriesByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}

	dtoCategories := make([]dto.Category, 0, len(categories))
	for _, cat := range categories {
		dtoCategories = append(dtoCategories, dto.Category{
			ID:     cat.ID,
			Name:   cat.Name,
			UserID: cat.UserID,
		})
	}
	return &dto.GetCategoriesResp{Categories: dtoCategories}, nil
}

func NewCategoryService(categoryDao *dao.CategoryDao) *CategoryService {
	return &CategoryService{categoryDao: categoryDao}
}
