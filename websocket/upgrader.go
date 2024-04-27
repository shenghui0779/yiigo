package websocket

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SetUpgrader 设置 websocket Upgrader
func SetUpgrader(up *websocket.Upgrader) {
	upgrader = up
}

// UpgradeConn websocket协议连接
type UpgradeConn struct {
	conn   *websocket.Conn
	authOK bool
	authFn func(ctx context.Context, msg *Message) (*Message, error)
}

// Read 读消息
func (c *UpgradeConn) Read(ctx context.Context, handler func(ctx context.Context, msg *Message) (*Message, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		t, b, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg *Message
		// if `authFunc` is not nil and unauthorized, need to authorize first.
		if c.authFn != nil && !c.authOK {
			msg, err = c.authFn(ctx, NewMessage(t, b))
			if err != nil {
				msg = TextMessage(err.Error())
			} else {
				c.authOK = true
			}
		} else {
			if handler != nil {
				msg, err = handler(ctx, NewMessage(t, b))
				if err != nil {
					msg = TextMessage(err.Error())
				}
			}
		}
		if msg != nil {
			c.conn.WriteMessage(msg.T(), msg.V())
		}
	}
}

// Write 写消息
func (c *UpgradeConn) Write(ctx context.Context, msg *Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// if `authFn` is not nil and unauthorized, disable to write message.
	if c.authFn != nil && !c.authOK {
		return errors.New("write msg disabled becaule of unauthorized")
	}

	return c.conn.WriteMessage(msg.T(), msg.V())
}

// Close 关闭连接
func (c *UpgradeConn) Close() error {
	return c.conn.Close()
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
func Upgrade(w http.ResponseWriter, r *http.Request, authFn func(ctx context.Context, msg *Message) (*Message, error)) (*UpgradeConn, error) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	conn := &UpgradeConn{
		conn:   c,
		authFn: authFn,
	}
	return conn, nil
}
