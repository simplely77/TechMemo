package service

import (
	"context"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
)

type UserService struct {
	userDao *dao.UserDao
}

func NewUserService(userDao *dao.UserDao) *UserService {
	return &UserService{userDao: userDao}
}

func (s *UserService) Register(ctx context.Context, username, password string) error {
	exists, err := s.userDao.CheckUserExists(ctx, username)
	if err != nil {
		return errors.InternalErr
	}

	if exists {
		return errors.UserExists
	}

	// TODO: 创建用户（Insert）
	return nil
}
