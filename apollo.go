package yiigo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/philchia/agollo"
)

// cfgM 配置映射存储
var cfgM sync.Map

// ApolloConfig 配置
type ApolloConfig interface {
	// Namespace return config namespace
	Namespace() string
	// DoExtra do some extra things with config
	DoExtra()
}

// DefaultApolloConfig default config
type DefaultApolloConfig struct {
	envkey    string `toml:"-"`
	namespace string `toml:"-"`
}

// Namespace returns namespace
func (c *DefaultApolloConfig) Namespace() string {
	return Env(c.envkey).String(c.namespace)
}

// DoExtra do some extra things with config
func (c *DefaultApolloConfig) DoExtra() {}

// NewDefaultConfig returns a default apollo config
func NewDefaultConfig(namespaceKey, defaultNamespace string) *DefaultApolloConfig {
	return &DefaultApolloConfig{
		envkey:    fmt.Sprintf("apollo.namespace.%s", namespaceKey),
		namespace: defaultNamespace,
	}
}

type apollo struct {
	AppID    string `toml:"appid"`
	Cluster  string `toml:"cluster"`
	Address  string `toml:"address"`
	CacheDir string `toml:"cache_dir"`
	isDev    bool   `toml:"-`
}

// Start start apollo
func (a *apollo) start(cfgs ...ApolloConfig) error {
	namespaces := make([]string, 0, len(cfgs))

	for _, cfg := range cfgs {
		if v := cfg.Namespace(); v != "" {
			namespaces = append(namespaces, v)
		}
	}

	conf := &agollo.Conf{
		AppID:          a.AppID,
		Cluster:        a.Cluster,
		IP:             a.Address,
		NameSpaceNames: namespaces,
		CacheDir:       a.CacheDir,
	}

	if err := agollo.StartWithConf(conf); err != nil {
		return err
	}

	a.registerConfig(cfgs...)

	go func() {
		events := agollo.WatchUpdate()

		for e := range events {
			if v, ok := cfgM.Load(e.Namespace); ok {
				a.updateConfigWithChanges(e, v.(ApolloConfig))
			}
		}
	}()

	logger.Info("yiigo: apollo is OK.")

	return nil
}

func (a *apollo) registerConfig(cfgs ...ApolloConfig) {
	for _, c := range cfgs {
		namespace := c.Namespace()

		a.setConfigWithNamespace(namespace, c)
		cfgM.Store(namespace, c)
	}
}

func (a *apollo) setConfigWithNamespace(namespace string, c ApolloConfig) {
	if a.isDev {
		if err := Env(namespace).Unmarshal(c); err == nil {
			return
		}
	}

	keys := agollo.GetAllKeys(namespace)

	l := len(keys)

	if l == 0 {
		Env(namespace).Unmarshal(c)

		return
	}

	m := make(map[string]string, l)

	for _, k := range keys {
		m[k] = agollo.GetStringValueWithNameSpace(namespace, k, "")
	}

	a.unmarshalConfig(c, m)

	c.DoExtra()
}

func (a *apollo) updateConfigWithChanges(e *agollo.ChangeEvent, c ApolloConfig) {
	if a.isDev {
		return
	}

	m := make(map[string]string, len(e.Changes))

	for k, v := range e.Changes {
		m[k] = v.NewValue
	}

	a.unmarshalConfig(c, m)

	c.DoExtra()
}

func (a *apollo) unmarshalConfig(c ApolloConfig, m map[string]string) {
	rv := reflect.Indirect(reflect.ValueOf(c))

	fieldNum := rv.NumField()
	t := rv.Type()

	for i := 0; i < fieldNum; i++ {
		field := t.Field(i)

		v, ok := m[field.Tag.Get("toml")]

		if !ok {
			continue
		}

		cfgVal := strings.TrimSpace(v)

		switch field.Type.Kind() {
		case reflect.String:
			rv.Field(i).SetString(cfgVal)
		case reflect.Int:
			var n int64

			if cfgVal != "" {
				n, _ = strconv.ParseInt(cfgVal, 10, 0)
			}

			rv.Field(i).SetInt(n)
		case reflect.Uint:
			var n uint64

			if cfgVal != "" {
				n, _ = strconv.ParseUint(cfgVal, 10, 0)
			}

			rv.Field(i).SetUint(n)
		case reflect.Int8:
			var n int64

			if cfgVal != "" {
				n, _ = strconv.ParseInt(cfgVal, 10, 8)
			}

			rv.Field(i).SetInt(n)
		case reflect.Uint8:
			var n uint64

			if cfgVal != "" {
				n, _ = strconv.ParseUint(cfgVal, 10, 8)
			}

			rv.Field(i).SetUint(n)
		case reflect.Int16:
			var n int64

			if cfgVal != "" {
				n, _ = strconv.ParseInt(cfgVal, 10, 16)
			}

			rv.Field(i).SetInt(n)
		case reflect.Uint16:
			var n uint64

			if cfgVal != "" {
				n, _ = strconv.ParseUint(cfgVal, 10, 16)
			}

			rv.Field(i).SetUint(n)
		case reflect.Int32:
			var n int64

			if cfgVal != "" {
				n, _ = strconv.ParseInt(cfgVal, 10, 32)
			}

			rv.Field(i).SetInt(n)
		case reflect.Uint32:
			var n uint64

			if cfgVal != "" {
				n, _ = strconv.ParseUint(cfgVal, 10, 32)
			}

			rv.Field(i).SetUint(n)
		case reflect.Int64:
			var n int64

			if cfgVal != "" {
				n, _ = strconv.ParseInt(cfgVal, 10, 64)
			}

			rv.Field(i).SetInt(n)
		case reflect.Uint64:
			var n uint64

			if cfgVal != "" {
				n, _ = strconv.ParseUint(cfgVal, 10, 64)
			}

			rv.Field(i).SetUint(n)
		case reflect.Float64:
			var n float64

			if cfgVal != "" {
				n, _ = strconv.ParseFloat(cfgVal, 10)
			}

			rv.Field(i).SetFloat(n)
		case reflect.Bool:
			var b bool

			if cfgVal != "" {
				b, _ = strconv.ParseBool(cfgVal)
			}

			rv.Field(i).SetBool(b)
		case reflect.Slice:
			ss := make([]string, 0)

			if cfgVal != "" {
				ss = strings.Split(cfgVal, field.Tag.Get("sep"))
			}

			switch field.Type.Elem().Kind() {
			case reflect.String:
				rv.Field(i).Set(reflect.ValueOf(ss))
			case reflect.Int:
				ns := make([]int, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.Atoi(s)
					ns = append(ns, n)
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Uint:
				ns := make([]uint, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseUint(s, 10, 0)
					ns = append(ns, uint(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Int8:
				ns := make([]int8, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseInt(s, 10, 8)
					ns = append(ns, int8(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Uint8:
				ns := make([]uint8, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseUint(s, 10, 8)
					ns = append(ns, uint8(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Int16:
				ns := make([]int16, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseInt(s, 10, 16)
					ns = append(ns, int16(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Uint16:
				ns := make([]uint16, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseUint(s, 10, 16)
					ns = append(ns, uint16(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Int32:
				ns := make([]int32, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseInt(s, 10, 32)
					ns = append(ns, int32(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Uint32:
				ns := make([]uint32, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseUint(s, 10, 32)
					ns = append(ns, uint32(n))
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Int64:
				ns := make([]int64, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseInt(s, 10, 64)
					ns = append(ns, n)
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Uint64:
				ns := make([]uint64, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseUint(s, 10, 64)
					ns = append(ns, n)
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Float64:
				ns := make([]float64, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseFloat(s, 10)
					ns = append(ns, n)
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			case reflect.Bool:
				ns := make([]bool, 0, len(ss))

				for _, s := range ss {
					n, _ := strconv.ParseBool(s)
					ns = append(ns, n)
				}

				rv.Field(i).Set(reflect.ValueOf(ns))
			}
		}
	}
}

// StartApollo start apollo
func StartApollo(cfgs ...ApolloConfig) error {
	if len(cfgs) == 0 {
		return ErrConfigNil
	}

	isDev := false

	if Env("app.env").String("dev") == "dev" {
		isDev = true
	}

	a := &apollo{isDev: isDev}

	if err := Env("apollo").Unmarshal(a); err != nil {
		return err
	}

	return a.start(cfgs...)
}
