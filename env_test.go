package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_env_Int(t *testing.T) {
	assert.Equal(t, int64(100), Env("app.amount").Int())
}

func Test_env_Ints(t *testing.T) {
	assert.Equal(t, []int64{80, 81, 82}, Env("app.ports").Ints())
}

func Test_env_Float(t *testing.T) {
	assert.Equal(t, 50.6, Env("app.weight").Float())
}

func Test_env_Floats(t *testing.T) {
	assert.Equal(t, []float64{23.5, 46.7, 45.9}, Env("app.prices").Floats())
}

func Test_env_String(t *testing.T) {
	assert.Equal(t, "dev", Env("app.env").String())
}

func Test_env_Strings(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1", "192.168.1.1", "192.168.1.80"}, Env("app.hosts").Strings())
}

func Test_env_Bool(t *testing.T) {
	assert.Equal(t, true, Env("app.debug").Bool())
}

func Test_env_Time(t *testing.T) {
	assert.Equal(t, time.Date(2019, 7, 12, 13, 3, 19, 0, time.UTC), Env("app.birthday").Time("2006-01-02 15:04:05"))
}

func Test_env_Map(t *testing.T) {
	assert.Equal(t, X{
		"env":      "dev",
		"debug":    true,
		"birthday": "2019-07-12 13:03:19",
		"amount":   int64(100),
		"hosts":    []interface{}{"127.0.0.1", "192.168.1.1", "192.168.1.80"},
		"ports":    []interface{}{int64(80), int64(81), int64(82)},
		"weight":   50.6,
		"prices":   []interface{}{23.5, 46.7, 45.9},
	}, Env("app").Map())
}

func Test_env_Unmarshal(t *testing.T) {
	type App struct {
		Env      string    `toml:"env"`
		Debug    bool      `toml:"debug"`
		Birthday string    `toml:"birthday"`
		Amount   int       `toml:"amount"`
		Hosts    []string  `toml:"hosts"`
		Ports    []int     `toml:"ports"`
		Weight   float64   `toml:"weight"`
		Prices   []float64 `toml:"prices"`
	}

	result := new(App)

	assert.Nil(t, Env("app").Unmarshal(result))
	assert.Equal(t, &App{
		Env:      "dev",
		Debug:    true,
		Birthday: "2019-07-12 13:03:19",
		Amount:   100,
		Hosts:    []string{"127.0.0.1", "192.168.1.1", "192.168.1.80"},
		Ports:    []int{80, 81, 82},
		Weight:   50.6,
		Prices:   []float64{23.5, 46.7, 45.9},
	}, result)
}

var (
	builder SQLBuilder

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	LoadEnvFromBytes([]byte(`[app]
env = "dev"
debug = true
birthday = "2019-07-12 13:03:19"
amount = 100
hosts = ["127.0.0.1", "192.168.1.1", "192.168.1.80"]
ports = [80, 81, 82]
weight = 50.6
prices = [23.5, 46.7, 45.9]`))

	builder = NewSQLBuilder(MySQL)

	privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS8)

	m.Run()
}
