package service

import (
	"context"
	stderrors "errors"
	"techmemo/backend/common/errors"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"
	"techmemo/backend/utils"

	"gorm.io/gorm"
)

type UserService struct {
	userDao *dao.UserDao
}

func (s *UserService) ValidateUser(ctx context.Context, userID int64) error {
	_, err := s.userDao.GetUserByID(ctx, userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.UserNotFound
		}
		return errors.InternalErr
	}

	return nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*dto.ProfileResp, error) {
	user, err := s.userDao.GetUserByID(ctx, userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.UserNotFound
		}
		return nil, errors.InternalErr
	}
	return &dto.ProfileResp{UserID: user.ID, Username: user.Username, CreatedAt: user.CreatedAt}, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	user, err := s.userDao.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.UserNotFound
		}
		return nil, errors.InternalErr
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return nil, errors.PasswordErr
	}

	token, err := utils.GenerateToken(user.ID, user.Username, config.AppConfig.JWT.Secret, config.AppConfig.JWT.ExpireHour)
	if err != nil {
		return nil, errors.InternalErr
	}

	return &dto.LoginResp{Token: token, UserID: user.ID, Username: user.Username}, nil
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterReq) (*dto.RegisterResp, error) {
	exists, err := s.userDao.CheckUserExists(ctx, req.Username)
	if err != nil {
		return nil, errors.InternalErr
	}

	if exists {
		return nil, errors.UserExists
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.InternalErr
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
	}
	err = s.userDao.Create(ctx, user)
	if err != nil {
		return nil, errors.InternalErr
	}
	return &dto.RegisterResp{UserID: user.ID, Username: user.Username, CreatedAt: user.CreatedAt}, nil
}

func NewUserService(userDao *dao.UserDao) *UserService {
	return &UserService{userDao: userDao}
}
