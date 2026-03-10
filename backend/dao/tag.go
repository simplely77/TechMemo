package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"

	"gorm.io/gorm"
)

type TagDao struct {
	q *query.Query
}

func (t *TagDao) GetTagsByTagIDsAndUserID(ctx context.Context, tagIDs []int64, userID int64) ([]*model.Tag, error) {
	if len(tagIDs) == 0 {
		return []*model.Tag{}, nil
	}
	return t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.ID.In(tagIDs...)).
		Where(t.q.Tag.UserID.Eq(userID)).
		Find()
}

func (t *TagDao) GetTagsByTagIDs(ctx context.Context, tagIDs []int64) ([]*model.Tag, error) {
	if len(tagIDs) == 0 {
		return []*model.Tag{}, nil
	}
	return t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.ID.In(tagIDs...)).
		Find()
}

func (t *TagDao) DeleteTag(ctx context.Context, id int64) error {
	result, err := t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (t *TagDao) UpdateTag(ctx context.Context, id int64, name string) error {
	result, err := t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.ID.Eq(id)).
		Update(t.q.Tag.Name, name)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (t *TagDao) CheckTagExists(ctx context.Context, userID int64, name string) (bool, error) {
	count, err := t.q.Tag.
		WithContext(ctx).
		Where(
			t.q.Tag.UserID.Eq(userID),
			t.q.Tag.Name.Eq(name)).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (t *TagDao) CreateTag(ctx context.Context, tag *model.Tag) error {
	return t.q.Tag.
		WithContext(ctx).
		Create(tag)
}

func (t *TagDao) CountTags(ctx context.Context, userID int64, keyword string) (int64, error) {
	q := t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.UserID.Eq(userID))

	if keyword != "" {
		q = q.Where(t.q.Tag.Name.Like("%" + keyword + "%"))
	}

	return q.Count()
}

func (t *TagDao) GetTags(ctx context.Context, userID int64, keyword string, offset int64, size int64) ([]*model.Tag, error) {
	q := t.q.Tag.
		WithContext(ctx).
		Where(t.q.Tag.UserID.Eq(userID))

	if keyword != "" {
		q = q.Where(t.q.Tag.Name.Like("%" + keyword + "%"))
	}

	return q.
		Offset(int(offset)).
		Limit(int(size)).
		Find()
}

func NewTagDao(q *query.Query) *TagDao {
	return &TagDao{q: q}
}
