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

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, config.AppConfig.JWT.Secret, config.AppConfig.JWT.AccessTokenExpireHour)
	if err != nil {
		return nil, errors.InternalErr
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, config.AppConfig.JWT.Secret, config.AppConfig.JWT.RefreshTokenExpireDay)
	if err != nil {
		return nil, errors.InternalErr
	}

	return &dto.LoginResp{AccessToken: accessToken, RefreshToken: refreshToken, UserID: user.ID, Username: user.Username}, nil
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

func (s *UserService) RefreshToken(ctx context.Context, req dto.RefreshTokenReq) (*dto.RefreshTokenResp, error) {
	claims, err := utils.ParseToken(req.RefreshToken, config.AppConfig.JWT.Secret)
	if err != nil {
		return nil, errors.Unauthorized
	}

	// 验证token类型必须是refreshToken
	if claims.TokenType != utils.RefreshToken {
		return nil, errors.Unauthorized
	}

	// 验证用户是否存在
	user, err := s.userDao.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.UserNotFound
		}
		return nil, errors.InternalErr
	}

	// 生成新的accessToken和refreshToken
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Username, config.AppConfig.JWT.Secret, config.AppConfig.JWT.AccessTokenExpireHour)
	if err != nil {
		return nil, errors.InternalErr
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, config.AppConfig.JWT.Secret, config.AppConfig.JWT.RefreshTokenExpireDay)
	if err != nil {
		return nil, errors.InternalErr
	}

	return &dto.RefreshTokenResp{AccessToken: newAccessToken, RefreshToken: newRefreshToken}, nil
}

func NewUserService(userDao *dao.UserDao) *UserService {
	return &UserService{userDao: userDao}
}
