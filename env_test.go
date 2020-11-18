package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_env_String(t *testing.T) {
	assert.Equal(t, "dev", Env("app.env").String())
}

func Test_env_Strings(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1", "192.168.1.1", "192.168.1.80"}, Env("app.hosts").Strings())
}

func Test_env_Int(t *testing.T) {
	assert.Equal(t, 100, Env("app.amount").Int())
}

func Test_env_Ints(t *testing.T) {
	assert.Equal(t, []int{80, 81, 82}, Env("app.ports").Ints())
}

func Test_env_Uint(t *testing.T) {
	assert.Equal(t, uint(100), Env("app.amount").Uint())
}

func Test_env_Uints(t *testing.T) {
	assert.Equal(t, []uint{80, 81, 82}, Env("app.ports").Uints())
}

func Test_env_Int8(t *testing.T) {
	assert.Equal(t, int8(100), Env("app.amount").Int8())
}

func Test_env_Int8s(t *testing.T) {
	assert.Equal(t, []int8{80, 81, 82}, Env("app.ports").Int8s())
}

func Test_env_Uint8(t *testing.T) {
	assert.Equal(t, uint8(100), Env("app.amount").Uint8())
}

func Test_env_Uint8s(t *testing.T) {
	assert.Equal(t, []uint8{80, 81, 82}, Env("app.ports").Uint8s())
}

func Test_env_Int16(t *testing.T) {
	assert.Equal(t, int16(100), Env("app.amount").Int16())
}

func Test_env_Int16s(t *testing.T) {
	assert.Equal(t, []int16{80, 81, 82}, Env("app.ports").Int16s())
}

func Test_env_Uint16(t *testing.T) {
	assert.Equal(t, uint16(100), Env("app.amount").Uint16())
}

func Test_env_Uint16s(t *testing.T) {
	assert.Equal(t, []uint16{80, 81, 82}, Env("app.ports").Uint16s())
}

func Test_env_Int32(t *testing.T) {
	assert.Equal(t, int32(100), Env("app.amount").Int32())
}

func Test_env_Int32s(t *testing.T) {
	assert.Equal(t, []int32{80, 81, 82}, Env("app.ports").Int32s())
}

func Test_env_Uint32(t *testing.T) {
	assert.Equal(t, uint32(100), Env("app.amount").Uint32())
}

func Test_env_Uint32s(t *testing.T) {
	assert.Equal(t, []uint32{80, 81, 82}, Env("app.ports").Uint32s())
}

func Test_env_Int64(t *testing.T) {
	assert.Equal(t, int64(100), Env("app.amount").Int64())
}

func Test_env_Int64s(t *testing.T) {
	assert.Equal(t, []int64{80, 81, 82}, Env("app.ports").Int64s())
}

func Test_env_Uint64(t *testing.T) {
	assert.Equal(t, uint64(100), Env("app.amount").Uint64())
}

func Test_env_Uint64s(t *testing.T) {
	assert.Equal(t, []uint64{80, 81, 82}, Env("app.ports").Uint64s())
}

func Test_env_Float64(t *testing.T) {
	assert.Equal(t, 50.6, Env("app.weight").Float64())
}

func Test_env_Float64s(t *testing.T) {
	assert.Equal(t, []float64{23.5, 46.7, 45.9}, Env("app.prices").Float64s())
}

func Test_env_Bool(t *testing.T) {
	assert.Equal(t, true, Env("app.debug").Bool())
}

func Test_env_Time(t *testing.T) {
	assert.Equal(t, time.Date(2016, 3, 19, 15, 3, 19, 0, time.UTC), Env("app.time").Time("2006-01-02 15:04:05"))
}

func Test_env_Map(t *testing.T) {
	assert.Equal(t, map[string]interface{}{
		"env":    "dev",
		"debug":  true,
		"time":   "2016-03-19 15:03:19",
		"amount": int64(100),
		"hosts":  []interface{}{"127.0.0.1", "192.168.1.1", "192.168.1.80"},
		"ports":  []interface{}{int64(80), int64(81), int64(82)},
		"weight": 50.6,
		"prices": []interface{}{23.5, 46.7, 45.9},
	}, Env("app").Map())
}

func Test_env_Unmarshal(t *testing.T) {
	type App struct {
		Env    string    `toml:"env"`
		Debug  bool      `toml:"debug"`
		Time   string    `toml:"time"`
		Amount int       `toml:"amount"`
		Hosts  []string  `toml:"hosts"`
		Ports  []int     `toml:"ports"`
		Weight float64   `toml:"weight"`
		Prices []float64 `toml:"prices"`
	}

	assert.Nil(t, Env("app").Unmarshal(&App{}))
}

var (
	builder *SQLBuilder

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	LoadEnvFromBytes([]byte(`[app]
env = "dev"
debug = true
time = "2016-03-19 15:03:19"
amount = 100
hosts = [ "127.0.0.1", "192.168.1.1", "192.168.1.80" ]
ports = [ 80, 81, 82 ]
weight = 50.6
prices = [ 23.5, 46.7, 45.9 ]`))

	builder = NewSQLBuilder(MySQL)

	privateKey, publicKey, _ = GenerateRSAKey(2048)

	m.Run()
}
