package response

import (
	"github.com/gin-gonic/gin"
)

func Response(ctx *gin.Context, httpStatus int, code int, data gin.H, msg string) {
	ctx.JSON(
		httpStatus,
		gin.H{
			"code": code,
			"data": data,
			"msg":  msg,
		})
}

func Success(ctx *gin.Context, data gin.H, msg string) {
	Response(ctx, 200, 200, data, msg)
}

func Fail(ctx *gin.Context, data gin.H, msg string) {
	Response(ctx, 400, 400, data, msg)
}

func ServeFail(ctx *gin.Context, data gin.H, msg string) {
	Response(ctx, 500, 500, data, msg)
}
