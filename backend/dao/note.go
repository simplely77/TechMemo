package dao

import (
	"context"
	"errors"
	"techmemo/backend/model"
	"techmemo/backend/query"
	"time"

	"gorm.io/gorm"
)

type NoteDao struct {
	q *query.Query
}

func (n *NoteDao) GetNotesByCid(ctx context.Context, cid int64) ([]*model.Note, error) {
	return n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.CategoryID.Eq(cid)).
		Where(n.q.Note.Status.Neq("deleted")).
		Find()
}

func (n *NoteDao) CountNotesByUid(ctx context.Context, userID int64) (int64, error) {
	return n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.UserID.Eq(userID)).
		Where(n.q.Note.Status.Neq("deleted")).
		Count()
}

func (n *NoteDao) GetNoteVersionByID(ctx context.Context, id int64) (*model.NoteVersion, error) {
	return n.q.NoteVersion.
		WithContext(ctx).
		Where(n.q.NoteVersion.ID.Eq(id)).
		First()
}

func (n *NoteDao) GetNoteVersions(ctx context.Context, noteID int64, sort string) ([]*model.NoteVersion, error) {
	q := n.q.NoteVersion.
		WithContext(ctx).
		Where(n.q.NoteVersion.NoteID.Eq(noteID))

	// 根据 sort 参数排序
	switch sort {
	case "created_at_asc":
		q = q.Order(n.q.NoteVersion.CreatedAt.Asc())
	default:
		q = q.Order(n.q.NoteVersion.CreatedAt.Desc())
	}

	return q.Find()
}

func (n *NoteDao) DeleteNoteByID(ctx context.Context, id int64) error {
	result, err := n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.ID.Eq(id)).
		Update(n.q.Note.Status, "deleted")
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (n *NoteDao) DeleteNoteTagsByNoteID(ctx context.Context, noteID int64) error {
	_, err := n.q.NoteTag.
		WithContext(ctx).
		Where(n.q.NoteTag.NoteID.Eq(noteID)).
		Delete()

	return err
}

func (n *NoteDao) UpdateNote(ctx context.Context, id int64, p UpdateNoteParams) error {
	updates := map[string]any{}

	if p.Title != nil {
		updates["title"] = *p.Title
	}
	if p.ContentMD != nil {
		updates["content_md"] = *p.ContentMD
	}
	if p.CategoryID != nil {
		updates["category_id"] = *p.CategoryID
	}
	if p.NoteType != nil {
		updates["note_type"] = *p.NoteType
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()

	_, err := n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.ID.Eq(id)).
		Updates(updates)

	return err
}

func (n *NoteDao) CheckNoteExistsByID(ctx context.Context, id int64) (bool, error) {
	count, err := n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.ID.Eq(id)).
		Where(n.q.Note.Status.Neq("deleted")).
		Count()
	return count > 0, err
}

func (n *NoteDao) GetNoteByID(ctx context.Context, id int64) (*model.Note, error) {
	note, err := n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.ID.Eq(id)).
		Where(n.q.Note.Status.Neq("deleted")).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return note, nil
}

func (n *NoteDao) GetNotesByIDs(ctx context.Context, ids []int64) ([]*model.Note, error) {
	return n.q.Note.
		WithContext(ctx).
		Where(n.q.Note.ID.In(ids...)).
		Where(n.q.Note.Status.Neq("deleted")).
		Find()
}

func (n *NoteDao) GetTagIDsByNoteID(ctx context.Context, noteID int64) ([]int64, error) {
	var tagIDs []int64
	err := n.q.NoteTag.
		WithContext(ctx).
		Where(n.q.NoteTag.NoteID.Eq(noteID)).
		Pluck(n.q.NoteTag.TagID, &tagIDs)
	return tagIDs, err
}

func (n *NoteDao) GetNotes(ctx context.Context, p GetNotesParams) ([]*model.Note, int64, error) {
	q := n.q.Note.WithContext(ctx).
		Where(n.q.Note.UserID.Eq(p.UserID)).
		Where(n.q.Note.Status.Neq("deleted"))

	if p.CategoryID > 0 {
		q = q.Where(n.q.Note.CategoryID.Eq(p.CategoryID))
	}

	if p.NoteType != "" {
		q = q.Where(n.q.Note.NoteType.Eq(p.NoteType))
	}

	if p.Keyword != "" {
		like := "%" + p.Keyword + "%"
		q = q.Where(
			n.q.Note.Title.Like(like),
			n.q.Note.ContentMd.Like(like),
		)
	}

	if len(p.TagIDs) > 0 {
		q = q.
			Join(
				n.q.NoteTag,
				n.q.NoteTag.NoteID.EqCol(n.q.Note.ID),
			).
			Where(n.q.NoteTag.TagID.In(p.TagIDs...))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	switch p.Sort {
	case "created_at_asc":
		q = q.Order(n.q.Note.CreatedAt.Asc())
	default:
		q = q.Order(n.q.Note.CreatedAt.Desc())
	}

	// 分页
	notes, err := q.
		Limit(int(p.Limit)).
		Offset(int(p.Offset)).
		Find()

	return notes, total, err
}

func (n *NoteDao) CreateNoteVersion(ctx context.Context, noteVersion *model.NoteVersion) error {
	return n.q.NoteVersion.
		WithContext(ctx).
		Create(noteVersion)
}

func (n *NoteDao) CreateNoteAndTags(ctx context.Context, noteTags []*model.NoteTag) error {
	return n.q.NoteTag.
		WithContext(ctx).
		CreateInBatches(noteTags, len(noteTags))
}

func (n *NoteDao) CreateNote(ctx context.Context, note *model.Note) error {
	return n.q.Note.
		WithContext(ctx).
		Create(note)
}

func NewNoteDao(q *query.Query) *NoteDao {
	return &NoteDao{q: q}
}
