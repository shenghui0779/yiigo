package session

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/iiinsomnia/yiigo"
)

const gosessid = "GOSESSID"

var store *sessions.CookieStore

// Start start session
func Start() {
	store = sessions.NewCookieStore([]byte(yiigo.Env.String("session.secret", "N0awmAuS2OziVFu^9!*0LY7MeCRgQ&z0")))
}

// Get get session key - value
func Get(c *gin.Context, key string, defaultValule ...interface{}) (interface{}, error) {
	session, err := store.Get(c.Request, gosessid)

	if err != nil {
		return nil, err
	}

	// Get some session values.
	v, ok := session.Values[key]

	if !ok {
		if len(defaultValule) > 0 {
			return defaultValule[0], nil
		}

		return nil, nil
	}

	return v, nil
}

// Set set session key - value, duration: seconds
func Set(c *gin.Context, key string, data interface{}, duration ...int) error {
	session, err := store.Get(c.Request, gosessid)

	if err != nil {
		return err
	}

	if len(duration) > 0 {
		session.Options = &sessions.Options{
			Path:   "/",
			MaxAge: duration[0],
		}
	}

	// Set some session values.
	session.Values[key] = data
	// Save it before we write to the response/return from the handler.
	err = session.Save(c.Request, c.Writer)

	return err
}

// Delete delete session key
func Delete(c *gin.Context, key string) error {
	session, err := store.Get(c.Request, gosessid)

	if err != nil {
		return err
	}

	delete(session.Values, key)

	err = session.Save(c.Request, c.Writer)

	return err
}

// Destroy destroy session
func Destroy(c *gin.Context) error {
	session, err := store.Get(c.Request, gosessid)

	if err != nil {
		return err
	}

	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: -1,
	}

	err = session.Save(c.Request, c.Writer)

	return err
}
