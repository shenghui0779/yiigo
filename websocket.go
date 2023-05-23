package yiigo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var wsupgrader *websocket.Upgrader

// WSMsg websocket消息
type WSMsg interface {
	// T returns ws msg type.
	T() int

	// V returns ws msg value.
	V() []byte
}

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

// NewWSMsg 返回一个websocket消息
func NewWSMsg(t int, v []byte) WSMsg {
	return &wsmsg{
		t: t,
		v: v,
	}
}

// NewWSTextMsg 返回一个websocket纯文本消息
func NewWSTextMsg(v []byte) WSMsg {
	return &wsmsg{
		t: websocket.TextMessage,
		v: v,
	}
}

// NewWSBinaryMsg 返回一个websocket字节消息
func NewWSBinaryMsg(v []byte) WSMsg {
	return &wsmsg{
		t: websocket.BinaryMessage,
		v: v,
	}
}

// WSHandler websocket消息处理方法
type WSHandler func(ctx context.Context, msg WSMsg) (WSMsg, error)

// WSConn websocket连接
type WSConn interface {
	// Read 读消息
	Read(ctx context.Context, callback WSHandler) error

	// Write 写消息
	Write(ctx context.Context, msg WSMsg) error

	// Close 关闭连接
	Close(ctx context.Context)
}

type wsconn struct {
	name   string
	conn   *websocket.Conn
	authOK bool
	authFn WSHandler
	log    func(ctx context.Context, v ...any)
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
					c.log(ctx, fmt.Sprintf("conn(%s) closed: %v", c.name, err))

					return nil
				}

				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					c.log(ctx, fmt.Sprintf("conn(%s) closed unexpectedly: %v", c.name, err))

					return nil
				}

				return err
			}

			var msg WSMsg

			// if `authFunc` is not nil and unauthorized, need to authorize first.
			if c.authFn != nil && !c.authOK {
				msg, err = c.authFn(ctx, NewWSMsg(t, b))

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
				if err = c.conn.WriteMessage(msg.T(), msg.V()); err != nil {
					c.log(ctx, fmt.Sprintf("conn(%s) write msg failed, got err: %v", c.name, err))
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

	// if `authFn` is not nil and unauthorized, disable to write message.
	if c.authFn != nil && !c.authOK {
		c.log(ctx, fmt.Sprintf("conn(%s) write msg disabled due to unauthorized", c.name))

		return nil
	}

	if err := c.conn.WriteMessage(msg.T(), msg.V()); err != nil {
		return err
	}

	return nil
}

func (c *wsconn) Close(ctx context.Context) {
	if err := c.conn.Close(); err != nil {
		c.log(ctx, fmt.Sprintf("close conn(%s) failed, got err: %v", c.name, err))
	}
}

// WSOption websocket连接选项
type WSOption func(c *wsconn)

// WithWSAuth 设置授权方法
func WithWSAuth(fn WSHandler) WSOption {
	return func(c *wsconn) {
		c.authFn = fn
	}
}

// WithWSLogger 设置日志
func WithWSLogger(fn func(ctx context.Context, v ...any)) WSOption {
	return func(c *wsconn) {
		c.log = fn
	}
}

// NewWSConn 生成一个websocket连接
func NewWSConn(name string, w http.ResponseWriter, r *http.Request, options ...WSOption) (WSConn, error) {
	if wsupgrader == nil {
		return nil, errors.New("upgrader is nil (forgotten configure?)")
	}

	c, err := wsupgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	conn := &wsconn{
		name: name,
		conn: c,
		log: func(ctx context.Context, v ...any) {
			logger.Error("err websocket", zap.String("err", fmt.Sprint(v...)))
		},
	}

	for _, f := range options {
		f(conn)
	}

	return conn, nil
}
