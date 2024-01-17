package yiigo

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader *websocket.Upgrader

// WSMessage websocket消息
type WSMessage struct {
	t int
	v []byte
}

func (m *WSMessage) T() int {
	return m.t
}

func (m *WSMessage) V() []byte {
	return m.v
}

// NewWSMessage 返回一个websocket消息
func NewWSMessage(t int, v []byte) *WSMessage {
	return &WSMessage{
		t: t,
		v: v,
	}
}

// NewWSTextMsg 返回一个websocket.TextMessage
func NewWSTextMsg(s string) *WSMessage {
	return &WSMessage{
		t: websocket.TextMessage,
		v: []byte(s),
	}
}

// NewWSBinaryMsg 返回一个websocket.BinaryMessage
func NewWSBinaryMsg(v []byte) *WSMessage {
	return &WSMessage{
		t: websocket.BinaryMessage,
		v: v,
	}
}

// WSConn websocket连接
type WSConn interface {
	// Read 读消息
	Read(ctx context.Context, handler func(ctx context.Context, msg *WSMessage) (*WSMessage, error)) error
	// Write 写消息
	Write(ctx context.Context, msg *WSMessage) error
	// Close 关闭连接
	Close(ctx context.Context) error
}

type wsconn struct {
	conn   *websocket.Conn
	authOK bool
	authFn func(ctx context.Context, msg *WSMessage) (*WSMessage, error)
}

func (c *wsconn) Read(ctx context.Context, handler func(ctx context.Context, msg *WSMessage) (*WSMessage, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t, b, err := c.conn.ReadMessage()
			if err != nil {
				return err
			}

			var msg *WSMessage

			// if `authFunc` is not nil and unauthorized, need to authorize first.
			if c.authFn != nil && !c.authOK {
				msg, err = c.authFn(ctx, NewWSMessage(t, b))
				if err != nil {
					msg = NewWSTextMsg(err.Error())
				} else {
					c.authOK = true
				}
			} else {
				if handler != nil {
					msg, err = handler(ctx, NewWSMessage(t, b))
					if err != nil {
						msg = NewWSTextMsg(err.Error())
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

func (c *wsconn) Write(ctx context.Context, msg *WSMessage) error {
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

// WSUpgrade upgrades the HTTP server connection to the WebSocket protocol.
func WSUpgrade(w http.ResponseWriter, r *http.Request, authFn func(ctx context.Context, msg *WSMessage) (*WSMessage, error)) (WSConn, error) {
	if upgrader == nil {
		return nil, errors.New("upgrader is nil (forgotten configure?)")
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	conn := &wsconn{
		conn:   c,
		authFn: authFn,
	}

	return conn, nil
}
