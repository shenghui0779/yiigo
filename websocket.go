package yiigo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

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

func NewWSMessage(t int, b []byte) *WSMessage {
	return &WSMessage{
		t: t,
		v: b,
	}
}

func WSTextMessage(s string) *WSMessage {
	return &WSMessage{
		t: websocket.TextMessage,
		v: []byte(s),
	}
}

func WSBinaryMessage(b []byte) *WSMessage {
	return &WSMessage{
		t: websocket.BinaryMessage,
		v: b,
	}
}

func WSCloseMessage(code int, text string) *WSMessage {
	return &WSMessage{
		t: websocket.PingMessage,
		v: websocket.FormatCloseMessage(code, text),
	}
}

func WSPingMessage(b []byte) *WSMessage {
	return &WSMessage{
		t: websocket.PingMessage,
		v: b,
	}
}

func WSPongMessage(b []byte) *WSMessage {
	return &WSMessage{
		t: websocket.PongMessage,
		v: b,
	}
}

// ----------------------------------------- dialer -----------------------------------------

var dialer = websocket.DefaultDialer

// SetDialer 设置 websocket dialer
func SetDialer(d *websocket.Dialer) {
	dialer = d
}

// DialConn websocket拨号连接，支持(读/写)失败重连
type DialConn struct {
	key    string
	addr   string
	header http.Header
	conn   *websocket.Conn
	mutex  singleflight.Group
}

func (c *DialConn) reconnect() error {
	// 并发安全，多次请求只会重连一次
	ch := c.mutex.DoChan(c.key, func() (interface{}, error) {
		// 创建新连接
		conn, _, err := dialer.Dial(c.addr, c.header)
		if err != nil {
			return false, err
		}

		// 关闭原连接
		c.conn.Close()
		// 设置新连接
		c.conn = conn
		// 5秒后清除缓存
		go func() {
			time.Sleep(5 * time.Second)
			c.mutex.Forget(c.key)
		}()

		return true, nil
	})

	// 设置超时，以防阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	select {
	case <-ctx.Done(): // 超时
		err = ctx.Err()
	case r := <-ch:
		err = r.Err
	}
	if err != nil {
		// 重连失败，清除缓存
		c.mutex.Forget(c.key)
	}

	return err
}

// ReadMessage 读取消息，若失败会尝试重连 (reconnectTimeout<=0 表示重连不超时)
func (c *DialConn) Read(reconnectTimeout time.Duration, handler func(msg *WSMessage)) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("websocket read panic", zap.Any("error", err), zap.ByteString("stack", debug.Stack()))
		}
	}()

	for {
		t, b, err := c.conn.ReadMessage()
		if err == nil {
			handler(NewWSMessage(t, b))
			continue
		}

		var cancel context.CancelFunc

		ctx := context.Background()
		if reconnectTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, reconnectTimeout)
		}

		// 尝试重连
		for {
			select {
			case <-ctx.Done():
				if cancel != nil {
					cancel()
				}
				return fmt.Errorf("websocket reconnect timeout: %w", ctx.Err())
			default:
			}

			// 1秒后尝试重连
			time.Sleep(time.Second)
			if err = c.reconnect(); err == nil {
				if cancel != nil {
					cancel()
				}
				break
			}
		}
	}
}

// WriteMessage 写入消息，若失败会尝试重连 (reconnectTimeout<=0 表示重连不超时)
func (c *DialConn) Write(reconnectTimeout time.Duration, msg *WSMessage) error {
	for {
		err := c.conn.WriteMessage(msg.T(), msg.V())
		if err == nil {
			return nil
		}

		var cancel context.CancelFunc

		ctx := context.Background()
		if reconnectTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, reconnectTimeout)
		}

		// 尝试重连
		for {
			select {
			case <-ctx.Done():
				if cancel != nil {
					cancel()
				}
				return fmt.Errorf("websocket reconnect timeout: %w", ctx.Err())
			default:
			}

			// 1秒后尝试重连
			time.Sleep(time.Second)
			if err = c.reconnect(); err == nil {
				if cancel != nil {
					cancel()
				}
				break
			}
		}
	}
}

func (c *DialConn) Close() error {
	return c.conn.Close()
}

// WSDial 创建一个websocket拨号连接
func WSDial(ctx context.Context, addr string, header http.Header) (*DialConn, error) {
	conn, _, err := dialer.DialContext(ctx, addr, header)
	if err != nil {
		return nil, err
	}

	return &DialConn{
		key:    MD5(addr),
		addr:   addr,
		header: header,
		conn:   conn,
	}, nil
}

// ----------------------------------------- upgrader -----------------------------------------

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

type UpgradeConn struct {
	conn   *websocket.Conn
	authOK bool
	authFn func(ctx context.Context, msg *WSMessage) (*WSMessage, error)
}

func (c *UpgradeConn) Read(ctx context.Context, handler func(ctx context.Context, msg *WSMessage) (*WSMessage, error)) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("websocket read panic", zap.Any("error", err), zap.ByteString("stack", debug.Stack()))
		}
	}()

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

		var msg *WSMessage

		// if `authFunc` is not nil and unauthorized, need to authorize first.
		if c.authFn != nil && !c.authOK {
			msg, err = c.authFn(ctx, NewWSMessage(t, b))
			if err != nil {
				msg = WSTextMessage(err.Error())
			} else {
				c.authOK = true
			}
		} else {
			if handler != nil {
				msg, err = handler(ctx, NewWSMessage(t, b))
				if err != nil {
					msg = WSTextMessage(err.Error())
				}
			}
		}

		if msg != nil {
			c.conn.WriteMessage(msg.T(), msg.V())
		}
	}
}

func (c *UpgradeConn) Write(ctx context.Context, msg *WSMessage) error {
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

func (c *UpgradeConn) Close() error {
	return c.conn.Close()
}

// WSUpgrade upgrades the HTTP server connection to the WebSocket protocol.
func WSUpgrade(w http.ResponseWriter, r *http.Request, authFn func(ctx context.Context, msg *WSMessage) (*WSMessage, error)) (*UpgradeConn, error) {
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
