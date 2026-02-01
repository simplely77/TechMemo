package dao

import (
	"context"
	"techmemo/backend/query"
)

type UserDao struct {
	q *query.Query
}

func NewUserDao(q *query.Query) *UserDao {
	return &UserDao{q: q}
}

func (d *UserDao) CheckUserExists(ctx context.Context, username string) (bool, error) {
	user := d.q.User

	count, err := user.
		WithContext(ctx).
		Where(user.Username.Eq(username)).
		Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
