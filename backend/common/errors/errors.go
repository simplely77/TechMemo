package errors

type ErrorCode struct {
	Code    int
	Message string
}

// Error implements [error].
func (e ErrorCode) Error() string {
	panic("unimplemented")
}

var (
	Success = ErrorCode{20000, "success"}

	// 通用错误
	InvalidParam = ErrorCode{40001, "参数错误"}
	Unauthorized = ErrorCode{40002, "未登录"}
	Forbidden    = ErrorCode{40003, "无权限"}

	// 用户模块
	UserNotFound = ErrorCode{41001, "用户不存在"}
	UserExists   = ErrorCode{41002, "用户已存在"}
	PasswordErr  = ErrorCode{41003, "密码错误"}

	// 系统错误
	InternalErr = ErrorCode{50001, "服务器内部错误"}
)
