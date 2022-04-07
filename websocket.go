package yiigo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tidwall/pretty"
	"go.uber.org/zap"
)

var wsupgrader *websocket.Upgrader

// WSMsg websocket message
type WSMsg interface {
	// T returns ws msg type.
	T() int

	// V returns ws msg value.
	V() []byte
}

// wsmsg websocket message
type wsmsg struct {
	t int
	v []byte
}

func (m *wsmsg) T() int {
	return m.t
}

func (m *wsmsg) V() []byte {
	return m.v
}

// NewWSMsg returns a new ws message.
func NewWSMsg(t int, v []byte) WSMsg {
	return &wsmsg{
		t: t,
		v: v,
	}
}

// NewWSTextMsg returns a new ws text message.
func NewWSTextMsg(v []byte) WSMsg {
	return &wsmsg{
		t: websocket.TextMessage,
		v: v,
	}
}

// NewWSBinaryMsg returns a new ws binary message.
func NewWSBinaryMsg(v []byte) WSMsg {
	return &wsmsg{
		t: websocket.BinaryMessage,
		v: v,
	}
}

// WSHandler the function to handle ws message.
type WSHandler func(ctx context.Context, msg WSMsg) (WSMsg, error)

// WSConn websocket connection
type WSConn interface {
	// Read reads message from ws connection.
	Read(ctx context.Context, callback WSHandler) error

	// Write writes message to ws connection.
	Write(ctx context.Context, msg WSMsg) error

	// Close closes ws connection.
	Close(ctx context.Context)

	// AuthOK returns true if authorized success (authorization handler has specified).
	AuthOK(ctx context.Context) bool
}

type wsconn struct {
	name     string
	conn     *websocket.Conn
	authOK   bool
	authFunc WSHandler
	logger   CtxLogger
}

func (c *wsconn) Read(ctx context.Context, callback WSHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t, b, err := c.conn.ReadMessage()

			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					c.logger.Info(ctx, fmt.Sprintf("conn(%s) closed", c.name), zap.String("msg", err.Error()))

					return nil
				}

				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					c.logger.Warn(ctx, fmt.Sprintf("conn(%s) closed unexpectedly", c.name), zap.String("msg", err.Error()))

					return nil
				}

				return err
			}

			c.logger.Info(ctx, fmt.Sprintf("conn(%s) read msg", c.name), zap.Int("msg.T", t), zap.ByteString("msg.V", pretty.Ugly(b)))

			var msg WSMsg

			// if `authFunc` is not nil and unauthorized, need to authorize first.
			if c.authFunc != nil && !c.authOK {
				msg, err = c.authFunc(ctx, NewWSMsg(t, b))

				if err != nil {
					msg = NewWSTextMsg([]byte(err.Error()))
				} else {
					c.authOK = true
				}
			} else {
				if callback != nil {
					msg, err = callback(ctx, NewWSMsg(t, b))

					if err != nil {
						msg = NewWSTextMsg([]byte(err.Error()))
					}
				}
			}

			if msg != nil {
				c.logger.Info(ctx, fmt.Sprintf("conn(%s) write msg", c.name), zap.Int("msg.T", msg.T()), zap.ByteString("msg.V", pretty.Ugly(msg.V())))

				if err = c.conn.WriteMessage(msg.T(), msg.V()); err != nil {
					c.logger.Err(ctx, fmt.Sprintf("err conn(%s) write msg", c.name), zap.Error(err), zap.Int("msg.T", msg.T()), zap.ByteString("msg.V", pretty.Ugly(msg.V())))
				}
			}
		}
	}
}

func (c *wsconn) Write(ctx context.Context, msg WSMsg) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// if `authFunc` is not nil and unauthorized, disable to write message.
	if c.authFunc != nil && !c.authOK {
		c.logger.Warn(ctx, fmt.Sprintf("conn(%s) write msg disabled due to unauthorized", c.name), zap.Int("msg.T", msg.T()), zap.ByteString("msg.V", pretty.Ugly(msg.V())))

		return nil
	}

	c.logger.Info(ctx, fmt.Sprintf("conn(%s) write msg", c.name), zap.Int("msg.T", msg.T()), zap.ByteString("msg.V", pretty.Ugly(msg.V())))

	if err := c.conn.WriteMessage(msg.T(), msg.V()); err != nil {
		return err
	}

	return nil
}

func (c *wsconn) Close(ctx context.Context) {
	if err := c.conn.Close(); err != nil {
		c.logger.Err(ctx, fmt.Sprintf("err close conn(%s)", c.name), zap.Error(err))
	}
}

func (c *wsconn) AuthOK(ctx context.Context) bool {
	if c.authFunc == nil {
		c.logger.Warn(ctx, "authorization handler not specified")
	}

	return c.authOK
}

// WSOption ws connection option.
type WSOption func(c *wsconn)

// WithWSAuth specifies authorization for ws connection.
func WithWSAuth(fn WSHandler) WSOption {
	return func(c *wsconn) {
		c.authFunc = fn
	}
}

// WithWSLogger specifies the logger for ws connection.
func WithWSLogger(l CtxLogger) WSOption {
	return func(c *wsconn) {
		c.logger = l
	}
}

type wsLogger struct{}

func (l *wsLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Info(fmt.Sprintf("[ws] %s", msg), fields...)
}

func (l *wsLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Warn(fmt.Sprintf("[ws] %s", msg), fields...)
}

func (l *wsLogger) Err(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Error(fmt.Sprintf("[ws] %s", msg), fields...)
}

// NewWSConn returns a new ws connection.
func NewWSConn(name string, w http.ResponseWriter, r *http.Request, options ...WSOption) (WSConn, error) {
	if wsupgrader == nil {
		return nil, errors.New("upgrader is nil (forgotten configure?)")
	}

	c, err := wsupgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	conn := &wsconn{
		name:   name,
		conn:   c,
		logger: new(wsLogger),
	}

	for _, f := range options {
		f(conn)
	}

	return conn, nil
}
