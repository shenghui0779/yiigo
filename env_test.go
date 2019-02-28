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
// ports = [ 50001, 50002, 50003 ]
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
				defaultValue: []int{10086},
			},
			want: []int{50001, 50002, 50003},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := Env.Ints(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Ints() = %v, want %v", got, tt.want)
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
				defaultValue: []int64{10086},
			},
			want: []int64{50001, 50002, 50003},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Env.Int64s(tt.args.key, tt.args.defaultValue...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("env.Int64s() = %v, want %v", got, tt.want)
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
				"ports":  []interface{}{int64(50001), int64(50002), int64(50003)},
				"weight": 50.6,
				"prices": []interface{}{23.5, 46.7, 45.9},
			},
		},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Env.Unmarshal(tt.args.key, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("env.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
