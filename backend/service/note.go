package service

import (
	"context"
	stderrors "errors"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"
	"techmemo/backend/query"

	"gorm.io/gorm"
)

type NoteService struct {
	noteDao     *dao.NoteDao
	categoryDao *dao.CategoryDao
	tagDao      *dao.TagDao
	kpDao       *dao.KnowledgePointDao
	q           *query.Query
}

func (n *NoteService) CheckNoteOwnership(ctx context.Context, userID int64, noteID int64) bool {
	note, _ := n.noteDao.GetNoteByID(ctx, noteID)
	return note.UserID == userID
}

func (n *NoteService) RestoreNote(ctx context.Context, id int64, versionID int64) (*dto.GetNoteResp, error) {
	// 判断笔记是否存在
	exists, err := n.noteDao.CheckNoteExistsByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr
	}
	if !exists {
		return nil, errors.NoteNotFound
	}

	// 判断版本是否存在
	version, err := n.noteDao.GetNoteVersionByID(ctx, versionID)
	if err != nil {
		return nil, errors.InternalErr
	}
	if version == nil {
		return nil, errors.NoteVersionNotFound
	}

	// 事务：更新 note + 新建版本
	err = n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		// 更新笔记内容
		if err := noteDao.UpdateNote(ctx, id, dao.UpdateNoteParams{
			ContentMD: &version.ContentMd,
		}); err != nil {
			return err
		}

		// 新建一条历史版本（记录回档行为）
		if err := noteDao.CreateNoteVersion(ctx, &model.NoteVersion{
			NoteID:    id,
			ContentMd: version.ContentMd,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.InternalErr
	}

	// 返回更新后的笔记数据
	return n.GetNote(ctx, id)
}

func (n *NoteService) GetNoteVersions(ctx context.Context, id int64, sort string) (*dto.GetNoteVersionsResp, error) {
	// 判断笔记是否存在
	exists, err := n.noteDao.CheckNoteExistsByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr
	}
	if !exists {
		return nil, errors.NoteNotFound
	}
	// 获取数据库里的版本列表
	versions, err := n.noteDao.GetNoteVersions(ctx, id, sort)
	if err != nil {
		return nil, errors.InternalErr
	}

	// 转换成 DTO
	respVersions := make([]dto.Version, 0, len(versions))
	for _, v := range versions {
		respVersions = append(respVersions, dto.Version{
			ID:        v.ID,
			NoteID:    v.NoteID,
			ContentMD: v.ContentMd,
			CreatedAt: v.CreatedAt,
		})
	}

	return &dto.GetNoteVersionsResp{
		Versions: respVersions,
	}, nil
}

func (n *NoteService) DeleteNote(ctx context.Context, id int64) error {
	return n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		// 软删除笔记（修改 status 为 deleted）
		err := noteDao.DeleteNoteByID(ctx, id)
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return errors.NoteNotFound
			}
			return errors.InternalErr
		}

		// 同步删除关联知识点（含 knowledge_relation 和 embedding）
		kpDao := dao.NewKnowledgePointDao(tx)
		kps, err := kpDao.GetKnowledgePointsBySourceNoteID(ctx, id)
		if err != nil {
			return errors.InternalErr
		}
		for _, kp := range kps {
			if err := kpDao.DeleteKnowledgePoint(ctx, kp.ID); err != nil {
				return errors.InternalErr
			}
		}

		// 删除笔记的 embedding
		if _, err := tx.Embedding.WithContext(ctx).
			Where(tx.Embedding.TargetType.Eq("note")).
			Where(tx.Embedding.TargetID.Eq(id)).
			Delete(); err != nil {
			return errors.InternalErr
		}

		return nil
	})
}

// PermanentlyDeleteNote 永久删除笔记及其所有关联数据
func (n *NoteService) PermanentlyDeleteNote(ctx context.Context, id int64) error {
	return n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		// 检查笔记是否存在（包括已软删除的）
		note, err := tx.Note.WithContext(ctx).Where(tx.Note.ID.Eq(id)).First()
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return errors.NoteNotFound
			}
			return errors.InternalErr
		}

		// 只能永久删除已软删除的笔记
		if note.Status != "deleted" {
			return errors.InvalidParam
		}

		// 永久删除笔记及其所有关联数据
		if err := noteDao.PermanentlyDeleteNote(ctx, id); err != nil {
			return errors.InternalErr
		}

		return nil
	})
}

func (n *NoteService) UpdateNoteTags(ctx context.Context, userID int64, noteID int64, req *dto.UpdateNoteTagsReq) error {
	return n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		if err := noteDao.DeleteNoteTagsByNoteID(ctx, noteID); err != nil {
			return errors.InternalErr
		}

		if len(req.TagIDs) > 0 {
			tags, err := n.tagDao.GetTagsByTagIDsAndUserID(ctx, req.TagIDs, userID)
			if err != nil {
				return errors.InternalErr
			}
			if len(tags) != len(req.TagIDs) {
				return errors.InvalidParam
			}
			noteTags := make([]*model.NoteTag, 0, len(req.TagIDs))
			for _, tagID := range req.TagIDs {
				noteTags = append(noteTags, &model.NoteTag{
					NoteID: noteID,
					TagID:  tagID,
				})
			}

			if err := noteDao.CreateNoteAndTags(ctx, noteTags); err != nil {
				return errors.InternalErr
			}
		}

		return nil
	})
}

func (n *NoteService) UpdateNote(ctx context.Context, id int64, req *dto.UpdateNoteReq) (*dto.GetNoteResp, error) {
	err := n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		exists, err := noteDao.CheckNoteExistsByID(ctx, id)
		if err != nil {
			return errors.InternalErr
		}
		if !exists {
			return errors.NoteNotFound
		}

		// 只有当 ContentMD 不为空时才创建版本历史
		if req.ContentMD != nil {
			noteVersion := &model.NoteVersion{
				NoteID:    id,
				ContentMd: *req.ContentMD,
			}
			if err := noteDao.CreateNoteVersion(ctx, noteVersion); err != nil {
				return errors.InternalErr
			}
		}

		if err := noteDao.UpdateNote(ctx, id, dao.UpdateNoteParams{
			Title:      req.Title,
			ContentMD:  req.ContentMD,
			CategoryID: req.CategoryID,
		}); err != nil {
			return errors.InternalErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回更新后的笔记数据
	return n.GetNote(ctx, id)
}

func (n *NoteService) GetNote(ctx context.Context, id int64) (*dto.GetNoteResp, error) {
	note, err := n.noteDao.GetNoteByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr
	}
	if note == nil {
		return nil, errors.NoteNotFound
	}

	category, err := n.categoryDao.GetCategoryByID(ctx, note.CategoryID)
	if err != nil {
		return nil, errors.InternalErr
	}

	tagIDs, err := n.noteDao.GetTagIDsByNoteID(ctx, note.ID)
	if err != nil {
		return nil, errors.InternalErr
	}

	tags, err := n.tagDao.GetTagsByTagIDs(ctx, tagIDs)
	if err != nil {
		return nil, errors.InternalErr
	}

	respTags := make([]dto.NoteTag, 0, len(tags))
	for _, tag := range tags {
		respTags = append(respTags, dto.NoteTag{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	return &dto.GetNoteResp{
		ID:        note.ID,
		Title:     note.Title,
		ContentMD: note.ContentMd,
		NoteType:  note.NoteType,
		Category: dto.NoteCategory{
			ID:   category.ID,
			Name: category.Name,
		},
		Tags:      respTags,
		Status:    note.Status,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}, nil
}

func (n *NoteService) GetNotes(ctx context.Context, req *dto.GetNotesReq, userID int64) (*dto.GetNotesResp, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	notes, total, err := n.noteDao.GetNotes(ctx, dao.GetNotesParams{
		UserID:     userID,
		CategoryID: req.CategoryID,
		TagIDs:     req.TagIDs,
		Keyword:    req.Keyword,
		NoteType:   req.NoteType,
		Sort:       req.Sort,
		Limit:      req.PageSize,
		Offset:     offset,
	},
	)
	if err != nil {
		return nil, errors.InternalErr
	}

	respNotes := make([]*dto.NoteAbstract, 0, len(notes))
	for _, note := range notes {
		category, err := n.categoryDao.GetCategoryByID(ctx, note.CategoryID)
		if err != nil {
			return nil, errors.InternalErr
		}
		noteCategory := dto.NoteCategory{
			ID:   category.ID,
			Name: category.Name,
		}

		tagIDs, err := n.noteDao.GetTagIDsByNoteID(ctx, note.ID)
		if err != nil {
			return nil, errors.InternalErr
		}

		tags, err := n.tagDao.GetTagsByTagIDs(ctx, tagIDs)
		if err != nil {
			return nil, errors.InternalErr
		}

		noteTags := make([]dto.NoteTag, 0, len(tags))
		for _, tag := range tags {
			noteTags = append(noteTags, dto.NoteTag{
				ID:   tag.ID,
				Name: tag.Name,
			})
		}
		respNotes = append(respNotes, &dto.NoteAbstract{
			ID:        note.ID,
			Title:     note.Title,
			NoteType:  note.NoteType,
			Category:  noteCategory,
			Tags:      noteTags,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		})
	}

	return &dto.GetNotesResp{
		Notes:    respNotes,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (n *NoteService) CreateNoteWithTags(ctx context.Context, req *dto.CreateNoteReq, userID int64) (*dto.GetNoteResp, error) {
	exists, err := n.categoryDao.CheckCategoryByIDAndUserID(ctx, req.CategoryID, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	if !exists {
		return nil, errors.CategoryNotFound
	}
	if len(req.TagIDs) > 0 {
		tags, err := n.tagDao.GetTagsByTagIDsAndUserID(ctx, req.TagIDs, userID)
		if err != nil {
			return nil, errors.InternalErr
		}
		if len(tags) != len(req.TagIDs) {
			return nil, errors.TagNotFound
		}
	}
	var note *model.Note

	err = n.q.Transaction(func(tx *query.Query) error {
		noteDao := dao.NewNoteDao(tx)

		note = &model.Note{
			Title:      req.Title,
			ContentMd:  req.ContentMD,
			CategoryID: req.CategoryID,
			UserID:     userID,
		}

		if err := noteDao.CreateNote(ctx, note); err != nil {
			return errors.InternalErr
		}

		noteVersion := &model.NoteVersion{
			NoteID:    note.ID,
			ContentMd: note.ContentMd,
		}
		if err := noteDao.CreateNoteVersion(ctx, noteVersion); err != nil {
			return errors.InternalErr
		}

		if len(req.TagIDs) > 0 {
			noteTags := make([]*model.NoteTag, 0, len(req.TagIDs))
			for _, tagID := range req.TagIDs {
				noteTags = append(noteTags, &model.NoteTag{
					NoteID: note.ID,
					TagID:  tagID,
				})
			}

			if err := noteDao.CreateNoteAndTags(ctx, noteTags); err != nil {
				return errors.InternalErr
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return n.GetNote(ctx, note.ID)
}

func NewNoteService(noteDao *dao.NoteDao, categoryDao *dao.CategoryDao, tagDao *dao.TagDao, kpDao *dao.KnowledgePointDao, q *query.Query) *NoteService {
	return &NoteService{noteDao: noteDao, categoryDao: categoryDao, tagDao: tagDao, kpDao: kpDao, q: q}
}
