package websocket

import "github.com/gorilla/websocket"

// Message websocket消息
type Message struct {
	t int
	v []byte
}

func (m *Message) T() int {
	return m.t
}

func (m *Message) V() []byte {
	return m.v
}

func NewMessage(t int, b []byte) *Message {
	return &Message{
		t: t,
		v: b,
	}
}

func TextMessage(s string) *Message {
	return &Message{
		t: websocket.TextMessage,
		v: []byte(s),
	}
}

func BinaryMessage(b []byte) *Message {
	return &Message{
		t: websocket.BinaryMessage,
		v: b,
	}
}

func CloseMessage(code int, text string) *Message {
	return &Message{
		t: websocket.PingMessage,
		v: websocket.FormatCloseMessage(code, text),
	}
}

func PingMessage(b []byte) *Message {
	return &Message{
		t: websocket.PingMessage,
		v: b,
	}
}

func PongMessage(b []byte) *Message {
	return &Message{
		t: websocket.PongMessage,
		v: b,
	}
}
