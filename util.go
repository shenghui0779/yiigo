package yiigo

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// X is a convenient alias for a map[string]interface{} map
type X map[string]interface{}

// MD5 获取字符串md5值
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsXhr 判断是否为Ajax请求
func IsXhr(c *gin.Context) bool {
	x := c.Request.Header.Get("X-Requested-With")

	if strings.ToLower(x) == "xmlhttprequest" {
		return true
	}

	return false
}

// Date 时间戳格式化日期
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}

// OK API返回成功
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

// Error API返回失败
func Error(c *gin.Context, code int, msg ...string) {
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
