# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GitHub release](https://img.shields.io/github/release/IIInsomnia/yiigo.svg)](https://github.com/iiinsomnia/yiigo/releases/latest)
[![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/iiinsomnia/yiigo/v4)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

Go 轻量级开发通用库

## Features

- Support [MySQL](https://github.com/go-sql-driver/mysql)
- Support [PostgreSQL](https://github.com/lib/pq)
- Support [MongoDB](https://github.com/mongodb/mongo-go-driver)
- Support [Redis](https://github.com/gomodule/redigo)
- Support [Zipkin](https://github.com/openzipkin/zipkin-go)
- Support [Apollo](https://github.com/philchia/agollo)
- Use [gomail](https://github.com/go-gomail/gomail) for email sending
- Use [toml](https://github.com/pelletier/go-toml) for configuration
- Use [sqlx](https://github.com/jmoiron/sqlx) for SQL executing
- Use [gorm](https://gorm.io/) for ORM operating
- Use [zap](https://github.com/uber-go/zap) for logging
- Many helpful methods, eg: http、cypto etc.

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
cache_dir = "./"

    [apollo.namespace]
    # 指定配置对应的namespace名称，用于在不同环境下（如：灰度和生产环境）指定不同的namespace

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
host = "smtp.exmail.qq.com"
port = 25
username = ""
password = ""
```

- config usage

```go
yiigo.Env("app.env").String()
yiigo.Env("app.debug").Bool()
```

#### Apollo

```go
type QiniuConfig struct {
    BucketName string `toml:"bucket_name"`
}

func (c *QiniuConfig) Namespace() string {
    return yiigo.ApolloNamespace("qiniu", "qiniu")
}

var Qiniu = new(QiniuConfig)

if err := yiigo.StartApollo(Qiniu); err != nil {
    log.Fatal(err)
}

fmt.Println(Qiniu.BucketName)
```

> Notice! Namespace read from `yiigo.toml` when `app.debug = true`

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
