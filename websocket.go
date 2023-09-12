package yiigo

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
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
	Close(ctx context.Context) error
}

type wsconn struct {
	conn   *websocket.Conn
	authOK bool
	authFn WSHandler
}

func (c *wsconn) Read(ctx context.Context, callback WSHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t, b, err := c.conn.ReadMessage()
			if err != nil {
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
					return err
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
		return errors.New("write msg disabled due to unauthorized")
	}

	return c.conn.WriteMessage(msg.T(), msg.V())
}

func (c *wsconn) Close(ctx context.Context) error {
	return c.conn.Close()
}

// NewWSConn 生成一个websocket连接
func NewWSConn(w http.ResponseWriter, r *http.Request, authFn WSHandler) (WSConn, error) {
	if wsupgrader == nil {
		return nil, errors.New("upgrader is nil (forgotten configure?)")
	}

	c, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	conn := &wsconn{
		conn:   c,
		authFn: authFn,
	}

	return conn, nil
}
