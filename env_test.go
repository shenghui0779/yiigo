package yiigo

import (
	"reflect"
	"testing"
	"time"
)

// env.toml
//
// [app]
// env = "dev"
// debug = true
// time = "2016-03-19 15:03:19"
// amount = 100
// hosts = [ "127.0.0.1", "192.168.1.1", "192.168.1.80" ]
// ports = [ 80, 81, 82 ]
// weight = 50.6
// prices = [ 23.5, 46.7, 45.9 ]

func Test_env_String(t *testing.T) {
	type args struct {
		key          string
		defaultValue []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				key:          "app.env",
				defaultValue: []string{"prod"},
			},
			want: "dev",
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.String(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Strings(t *testing.T) {
	type args struct {
		key          string
		defaultValue []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "t1",
			args: args{
				key:          "app.hosts",
				defaultValue: []string{"127.0.0.1"},
			},
			want: []string{"127.0.0.1", "192.168.1.1", "192.168.1.80"},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Strings(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Strings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []int{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Ints(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []int{88},
			},
			want: []int{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Ints(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Ints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []uint{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Uint(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Uint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uints(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint
	}
	tests := []struct {
		name string
		args args
		want []uint
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []uint{88},
			},
			want: []uint{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Uints(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Uints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int8(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int8
	}
	tests := []struct {
		name string
		args args
		want int8
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []int8{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int8(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Int8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int8s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int8
	}
	tests := []struct {
		name string
		args args
		want []int8
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []int8{88},
			},
			want: []int8{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Int8s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Int8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint8(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint8
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []uint8{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Uint8(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Uint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint8s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint8
	}
	tests := []struct {
		name string
		args args
		want []uint8
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []uint8{88},
			},
			want: []uint8{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Uint8s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Uint8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int16(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int16
	}
	tests := []struct {
		name string
		args args
		want int16
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []int16{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int16(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Int16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int16s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int16
	}
	tests := []struct {
		name string
		args args
		want []int16
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []int16{88},
			},
			want: []int16{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Int16s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Int16s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint16(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint16
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []uint16{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Uint16(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Uint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint16s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint16
	}
	tests := []struct {
		name string
		args args
		want []uint16
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []uint16{88},
			},
			want: []uint16{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Uint16s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Uint16s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int32(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int32
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []int32{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int32(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Int32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int32s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int32
	}
	tests := []struct {
		name string
		args args
		want []int32
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []int32{88},
			},
			want: []int32{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Int32s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Int32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint32(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []uint32{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Uint32(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Uint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint32s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint32
	}
	tests := []struct {
		name string
		args args
		want []uint32
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []uint32{88},
			},
			want: []uint32{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Uint32s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Uint32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int64(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []int64{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int64(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Int64s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []int64{88},
			},
			want: []int64{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int64s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Int64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint64(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.amount",
				defaultValue: []uint64{0},
			},
			want: 100,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Uint64(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Uint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Uint64s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.ports",
				defaultValue: []uint64{88},
			},
			want: []uint64{80, 81, 82},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Uint64s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Uint64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Float64(t *testing.T) {
	type args struct {
		key          string
		defaultValue []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.weight",
				defaultValue: []float64{0},
			},
			want: 50.6,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Float64(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Float64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Float64s(t *testing.T) {
	type args struct {
		key          string
		defaultValue []float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "t1",
			args: args{
				key:          "app.prices",
				defaultValue: []float64{0},
			},
			want: []float64{23.5, 46.7, 45.9},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Float64s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Float64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Bool(t *testing.T) {
	type args struct {
		key          string
		defaultValue []bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				key:          "app.debug",
				defaultValue: []bool{false},
			},
			want: true,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Bool(tt.args.key, tt.args.defaultValue...); got != tt.want {
				t.Errorf("env.Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Time(t *testing.T) {
	type args struct {
		key          string
		layout       string
		defaultValue []time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "t1",
			args: args{
				key:          "app.time",
				layout:       "2006-01-02 15:04:05",
				defaultValue: []time.Time{time.Now()},
			},
			want: time.Date(2016, 3, 19, 15, 3, 19, 0, time.UTC),
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Time(tt.args.key, tt.args.layout, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Map(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "t1",
			args: args{key: "app"},
			want: map[string]interface{}{
				"env":    "dev",
				"debug":  true,
				"time":   "2016-03-19 15:03:19",
				"amount": int64(100),
				"hosts":  []interface{}{"127.0.0.1", "192.168.1.1", "192.168.1.80"},
				"ports":  []interface{}{int64(80), int64(81), int64(82)},
				"weight": 50.6,
				"prices": []interface{}{23.5, 46.7, 45.9},
			},
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Map(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env_Unmarshal(t *testing.T) {
	type args struct {
		key  string
		dest interface{}
	}
	type App struct {
		Env    string    `toml:"env"`
		Debug  bool      `toml:"debug"`
		Time   string    `toml:"time"`
		Amount int       `toml:"amount"`
		Hosts  []string  `toml:"hosts"`
		Ports  []int     `toml:"ports"`
		Weight int       `toml:"weight"`
		Prices []float64 `toml:"prices"`
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				key:  "app",
				dest: &App{},
			},
			wantErr: false,
		},
	}

	if err := UseEnv("env.toml"); err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Env.Unmarshal(tt.args.key, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("env.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
