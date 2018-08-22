package yiigo

import (
	"crypto/md5"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// X is a convenient alias for a map[string]interface{}.
type X map[string]interface{}

// MD5 calculate the md5 hash of a string.
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Date format a local time/date and
// returns a string formatted according to the given format string using the given timestamp of int64.
// The default format is: 2006-01-02 15:04:05.
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}

// IsXhr checks if a request is xml-http-request (ajax).
func IsXhr(c *gin.Context) bool {
	x := c.Request.Header.Get("X-Requested-With")

	if strings.ToLower(x) == "xmlhttprequest" {
		return true
	}

	return false
}

// IP2long converts a string containing an (IPv4) Internet Protocol dotted address into a long integer
func IP2long(ip string) int64 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	ret := big.NewInt(0)
	ret.SetBytes(ipv4)

	return ret.Int64()
}

// Long2IP converts an long integer address into a string in (IPv4) Internet standard dotted format
func Long2IP(ip int64) string {
	ipv4 := fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))

	return ipv4
}

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
