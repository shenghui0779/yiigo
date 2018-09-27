package yiigo

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OK returns success of an API.
func OK(c *gin.Context, data ...interface{}) {
	obj := gin.H{
		"success": true,
		"code":    0,
		"msg":     "success",
	}

	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(http.StatusOK, obj)
}

// Err returns error of an API.
func Err(c *gin.Context, code int, msg ...string) {
	obj := gin.H{
		"success": false,
		"code":    code,
		"msg":     "something wrong",
	}

	if len(msg) > 0 {
		obj["msg"] = msg[0]
	}

	c.JSON(http.StatusOK, obj)
}
