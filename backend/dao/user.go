package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"
)

type UserDao struct {
	q *query.Query
}

func (d *UserDao) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return d.q.User.
		WithContext(ctx).
		Where(d.q.User.ID.Eq(userID)).
		First()
}

func NewUserDao(q *query.Query) *UserDao {
	return &UserDao{q: q}
}

func (d *UserDao) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return d.q.User.
		WithContext(ctx).
		Where(d.q.User.Username.Eq(username)).
		First()
}

func (d *UserDao) CheckUserExists(ctx context.Context, username string) (bool, error) {
	count, err := d.q.User.
		WithContext(ctx).
		Where(d.q.User.Username.Eq(username)).
		Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *UserDao) Create(ctx context.Context, user *model.User) error {
	return d.q.User.
		WithContext(ctx).
		Create(user)
}
