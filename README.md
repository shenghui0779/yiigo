# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GitHub release](https://img.shields.io/github/release/IIInsomnia/yiigo.svg)](https://github.com/iiinsomnia/yiigo/releases/latest)
[![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/iiinsomnia/yiigo/v4)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

Go 轻量级开发通用库

## Features

- 支持 [MySQL](https://github.com/go-sql-driver/mysql)
- 支持 [PostgreSQL](https://github.com/lib/pq)
- 支持 [MongoDB](https://github.com/mongodb/mongo-go-driver)
- 支持 [Redis](https://github.com/gomodule/redigo)
- 支持 [Zipkin](https://github.com/openzipkin/zipkin-go)
- 支持 [Apollo](https://github.com/philchia/agollo)
- 邮件使用 [gomail](https://github.com/go-gomail/gomail)
- 配置使用 [toml](https://github.com/pelletier/go-toml)
- SQL使用 [sqlx](https://github.com/jmoiron/sqlx)
- ORM使用 [gorm](https://gorm.io/)
- 日志使用 [zap](https://github.com/uber-go/zap)
- 包含很多帮助方法，如：http、cypto、date、IP 等

## Requirements

`Go1.11+`

## Installation

```sh
go get github.com/iiinsomnia/yiigo/v4
```

## Usage

#### Config

- `yiigo.toml`

```toml
[app]
env = "dev" # dev | beta | prod
debug = true

[apollo]
app_id = "test"
cluster = "default"
address = "127.0.0.1:8080"
namespace = ["apollo_test"]
cache_dir = "./"
accesskey_secret = ""
insecure_skip_verify = true

[db]

    [db.default]
    driver = "mysql"
    dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
    # dsn = "host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable" # pgsql
    max_open_conns = 20
    max_idle_conns = 10
    conn_max_lifetime = 60 # 秒

[mongo]

    [mongo.default]
    dsn = "mongodb://username:password@localhost:27017"
    connect_timeout = 10 # 秒
    pool_size = 10
    max_conn_idle_time = 60 # 秒
    mode = "primary" # primary | primary_preferred | secondary | secondary_preferred | nearest

[redis]

    [redis.default]
    address = "127.0.0.1:6379"
    password = ""
    database = 0
    connect_timeout = 10 # 秒
    read_timeout = 10 # 秒
    write_timeout = 10 # 秒
    pool_size = 10
    pool_limit = 20
    idle_timeout = 60 # 秒
    wait_timeout = 10 # 秒
    prefill_parallelism = 0

[log]

    [log.default]
    path = "app.log"
    max_size = 500
    max_age = 0
    max_backups = 0
    compress = true

[email]

    [email.default]
    host = "smtp.exmail.qq.com"
    port = 25
    username = ""
    password = ""

# apollo namespace
[apollo_test]
name = "yiigo"
```

- usage

```go
yiigo.Env("app.env").String("dev")
yiigo.Env("app.debug").Bool(true)
yiigo.Env("apollo_test.name").String("test")
```

> ⚠️注意！
>
> 如果配置了 `apollo`，则：
>
> 1. 配置优先从 `apollo` 读取，若不存在，则从 `yiigo.toml` 读取；
> 2. 若 `namespace` 不在 `apollo` 配置中，则其配置项从 `application` 中获取；
> 3. 当 `app.debug = true` 时，配置从 `yiigo.toml` 读取

#### MySQL

```go
// default db
yiigo.DB().Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
yiigo.Orm().First(&User{}, 1)

// other db
yiigo.DB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
yiigo.Orm("foo").First(&User{}, 1)
```

#### MongoDB

```go
// default mongodb
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

defer cancel()

yiigo.Mongo().Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})

// other mongodb
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

defer cancel()

yiigo.Mongo("foo").Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
```

#### Redis

```go
// default redis
conn, err := yiigo.Redis().Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis().Put(conn)

conn.Do("SET", "test_key", "hello world")

// other redis
conn, err := yiigo.Redis("foo").Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis("foo").Put(conn)

conn.Do("SET", "test_key", "hello world")
```

#### Zipkin

```go
reporter := yiigo.NewZipkinHTTPReporter("http://localhost:9411/api/v2/spans")

// sampler
sampler := zipkin.NewModuloSampler(1)
// endpoint
endpoint, _ := zipkin.NewEndpoint("yiigo-zipkin", "localhost")

tracer, err := yiigo.NewZipkinTracer(reporter,
    zipkin.WithLocalEndpoint(endpoint),
    zipkin.WithSharedSpans(false),
    zipkin.WithSampler(sampler),
)

if err != nil {
    log.Fatal(err)
}

client, err := tracer.HTTPClient(yiigo.WithZipkinClientOptions(zipkinHttp.ClientTrace(true)))

if err != nil {
    log.Fatal(err)
}

b, err := client.Get(context.Background(), "url...",
    yiigo.WithRequestHeader("Content-Type", "application/json; charset=utf-8"),
    yiigo.WithRequestTimeout(5*time.Second),
)

if err != nil {
    log.Fatal(err)
}

fmt.Println(string(b))
```

#### Logger

```go
// default logger
yiigo.Logger().Info("hello world")

// other logger
yiigo.Logger("foo").Info("hello world")
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/iiinsomnia/yiigo/v4)
- [TOML](https://github.com/toml-lang/toml)
- [Example](https://github.com/iiinsomnia/yiigo-example)

**Enjoy 😊**
