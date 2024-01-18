package websocket

import (
	"api/lib/hash"
	"api/lib/log"
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

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

// Read 读消息，若失败会尝试重连 (reconnectTimeout<=0 表示重连不超时)
func (c *DialConn) Read(reconnectTimeout time.Duration, handler func(msg *Message)) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error(context.Background(), "websocket read panic", zap.Any("error", err), zap.ByteString("stack", debug.Stack()))
		}
	}()

	for {
		t, b, err := c.conn.ReadMessage()
		if err == nil {
			handler(NewMessage(t, b))
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

// Write 写消息，若失败会尝试重连 (reconnectTimeout<=0 表示重连不超时)
func (c *DialConn) Write(reconnectTimeout time.Duration, msg *Message) error {
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

// Close 关闭连接
func (c *DialConn) Close() error {
	return c.conn.Close()
}

// Dial 创建一个websocket拨号连接
func Dial(ctx context.Context, addr string, header http.Header) (*DialConn, error) {
	conn, _, err := dialer.DialContext(ctx, addr, header)
	if err != nil {
		return nil, err
	}

	return &DialConn{
		key:    hash.MD5(addr),
		addr:   addr,
		header: header,
		conn:   conn,
	}, nil
}
