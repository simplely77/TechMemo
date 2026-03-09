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

type TagService struct {
	tagDao *dao.TagDao
}

func (t *TagService) DeleteTag(ctx context.Context, id int64) error {
	err := t.tagDao.DeleteTag(ctx, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.TagNotFound
		}
		return errors.InternalErr
	}
	return nil
}

func (t *TagService) UpdateTag(ctx context.Context, userID int64, id int64, name string) error {
	exists, err := t.tagDao.CheckTagExists(ctx, userID, name)
	if err != nil {
		return errors.InternalErr
	}

	if exists {
		return errors.TagExists
	}
	err = t.tagDao.UpdateTag(ctx, id, name)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.TagNotFound
		}
		return errors.InternalErr
	}
	return nil
}

func (t *TagService) CreateTag(ctx context.Context, userID int64, name string) error {
	exists, err := t.tagDao.CheckTagExists(ctx, userID, name)
	if err != nil {
		return errors.InternalErr
	}

	if exists {
		return errors.TagExists
	}

	tag := &model.Tag{
		Name:   name,
		UserID: userID,
	}

	err = t.tagDao.CreateTag(ctx, tag)
	if err != nil {
		return errors.InternalErr
	}

	return nil
}

func (t *TagService) GetTags(ctx context.Context, userID int64, keyword string, page int64, size int64) (*dto.GetTagsResp, error) {
	offset := (page - 1) * size

	tags, err := t.tagDao.GetTags(ctx, userID, keyword, offset, size)
	if err != nil {
		return nil, errors.InternalErr
	}

	total, err := t.tagDao.CountTags(ctx, userID, keyword)
	if err != nil {
		return nil, errors.InternalErr
	}

	dtoTags := make([]dto.Tag, 0, len(tags))
	for _, tag := range tags {
		dtoTags = append(dtoTags, dto.Tag{
			ID:     tag.ID,
			Name:   tag.Name,
			UserID: tag.UserID,
		})
	}

	return &dto.GetTagsResp{Tags: dtoTags, Total: total, Page: page, PageSize: size}, nil
}

func NewTagService(tagDao *dao.TagDao) *TagService {
	return &TagService{tagDao: tagDao}
}
