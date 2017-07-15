package yiigo

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// X is a convenient alias for a map[string]interface{} map
type X map[string]interface{}

// IsXhr 判断是否为Ajax请求
func IsXhr(c *gin.Context) bool {
	x := c.Request.Header.Get("X-Requested-With")

	if x == "XMLHttpRequest" {
		return true
	}

	return false
}

// ReturnSuccess API返回成功
func ReturnSuccess(c *gin.Context, data ...interface{}) {
	obj := gin.H{
		"code": 0,
		"msg":  "success",
	}

	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(http.StatusOK, obj)
}

// ReturnFailed API返回失败
func ReturnFailed(c *gin.Context, data ...interface{}) {
	obj := gin.H{
		"code": -1,
		"msg":  "failed",
	}

	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(http.StatusOK, obj)
}

// ReturnJSON API返回JSON数据
func ReturnJSON(c *gin.Context, code int, msg string, data ...interface{}) {
	obj := gin.H{
		"code": code,
		"msg":  msg,
	}

	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(http.StatusOK, obj)
}
