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
	NotFound     = ErrorCode{40004, "资源不存在"}

	// 用户模块
	UserNotFound = ErrorCode{41001, "用户不存在"}
	UserExists   = ErrorCode{41002, "用户已存在"}
	PasswordErr  = ErrorCode{41003, "密码错误"}

	// 分类模块
	CategoryExists   = ErrorCode{42001, "分类已存在"}
	CategoryNotFound = ErrorCode{42002, "分类不存在"}

	// 标签模块
	TagExists   = ErrorCode{43001, "标签已存在"}
	TagNotFound = ErrorCode{43002, "标签不存在"}

	//笔记模块
	NoteNotFound        = ErrorCode{44001, "笔记不存在"}
	NoteTagNotFound     = ErrorCode{44002, "笔记标签关系不存在"}
	NoteVersionNotFound = ErrorCode{44003, "笔记版本不存在"}

	// 系统错误
	InternalErr = ErrorCode{50001, "服务器内部错误"}
	DataBaseErr = ErrorCode{50002, "数据库操作错误"}
)
