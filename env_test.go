package yiigo

import "testing"

var (
	builder SQLBuilder

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	builder = NewMySQLBuilder(WithSQLDebug())

	privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS8)

	m.Run()
}
