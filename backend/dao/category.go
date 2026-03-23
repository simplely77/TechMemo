package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"

	"gorm.io/gorm"
)

type CategoryDao struct {
	q *query.Query
}

func (c *CategoryDao) CountCategoriesByUid(ctx context.Context, userID int64) (int64, error) {
	return c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.UserID.Eq(userID)).
		Count()
}

func (c *CategoryDao) CheckCategoryByIDAndUserID(ctx context.Context, id int64, userID int64) (bool, error) {
	count, err := c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.ID.Eq(id)).
		Where(c.q.Category.UserID.Eq(userID)).
		Count()
	return count > 0, err
}

func (c *CategoryDao) GetCategoryByID(ctx context.Context, id int64) (*model.Category, error) {
	return c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.ID.Eq(id)).
		First()
}

func (c *CategoryDao) DeleteCategory(ctx context.Context, id int64) error {
	result, err := c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *CategoryDao) UpdateCategory(ctx context.Context, id int64, name string) error {
	result, err := c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.ID.Eq(id)).
		Update(c.q.Category.Name, name)

	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *CategoryDao) CreateCategory(ctx context.Context, category *model.Category) error {
	return c.q.Category.
		WithContext(ctx).
		Create(category)
}

func (c *CategoryDao) CheckCategoryExists(ctx context.Context, userID int64, name string) (bool, error) {
	count, err := c.q.Category.
		WithContext(ctx).
		Where(
			c.q.Category.UserID.Eq(userID),
			c.q.Category.Name.Eq(name)).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *CategoryDao) GetCategoriesByUserID(ctx context.Context, userID int64) ([]*model.Category, error) {
	return c.q.Category.
		WithContext(ctx).
		Where(c.q.Category.UserID.Eq(userID)).
		Find()
}

func NewCategoryDao(q *query.Query) *CategoryDao {
	return &CategoryDao{q: q}
}
