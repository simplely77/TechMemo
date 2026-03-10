package response

import (
	"techmemo/backend/common/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    errors.Success.Code,
		Message: errors.Success.Message,
		Data:    data,
	})
}

func Fail(c *gin.Context, err errors.ErrorCode) {
	c.JSON(200, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

// FailErr 将 error 安全转换为 ErrorCode 后返回，若转换失败则返回 InternalErr。
func FailErr(c *gin.Context, err error) {
	if ec, ok := err.(errors.ErrorCode); ok {
		Fail(c, ec)
	} else {
		Fail(c, errors.InternalErr)
	}
}
