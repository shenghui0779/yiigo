# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/iiinsomnia/yiigo?status.svg)](https://godoc.org/github.com/iiinsomnia/yiigo)
[![GitHub release](https://img.shields.io/github/release/IIInsomnia/yiigo.svg)](https://github.com/iiinsomnia/yiigo/releases/latest)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

A simple and light library which makes Golang development easier !

## Features

- Support [MySQL](https://github.com/go-sql-driver/mysql)
- Support [PostgreSQL](https://github.com/lib/pq)
- Support [MongoDB](https://github.com/mongodb/mongo-go-driver)
- Support [Redis](https://github.com/gomodule/redigo)
- Support [Zipkin](https://github.com/openzipkin/zipkin-go)
- Use [gomail](https://github.com/go-gomail/gomail) for email sending
- Use [toml](https://github.com/pelletier/go-toml) for configuration
- Use [sqlx](https://github.com/jmoiron/sqlx) for SQL executing
- Use [zap](https://github.com/uber-go/zap) for logging

## Requirements

`Go1.11+`

## Installation

```sh
go get github.com/iiinsomnia/yiigo/v3
```

## Usage

#### MySQL

```go
// default db
yiigo.RegisterDB(yiigo.AsDefault, yiigo.MySQL, "root:root@tcp(localhost:3306)/test")

yiigo.DB.Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)

// other db
yiigo.RegisterDB("foo", yiigo.MySQL, "root:root@tcp(localhost:3306)/foo")

yiigo.UseDB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
```

#### MongoDB

```go
// default mongodb
yiigo.RegisterMongoDB(yiigo.AsDefault, "mongodb://localhost:27017")

ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
yiigo.Mongo.Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})

// other mongodb
yiigo.RegisterMongoDB("foo", "mongodb://localhost:27017")

ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
yiigo.UseMongo("foo").Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
```

#### Redis

```go
// default redis
yiigo.RegisterRedis(yiigo.AsDefault, "localhost:6379")

conn, err := yiigo.Redis.Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis.Put(conn)

conn.Do("SET", "test_key", "hello world")

// other redis
yiigo.RegisterRedis("foo", "localhost:6379")

foo := yiigo.UseRedis("foo")
conn, err := foo.Get()

if err != nil {
    log.Fatal(err)
}

defer foo.Put(conn)

conn.Do("SET", "test_key", "hello world")
```

#### Config

```go
// env.toml
//
// [app]
// env = "dev"
// debug = true
// port = 50001

yiigo.UseEnv("env.toml")

yiigo.Env.Bool("app.debug", true)
yiigo.Env.Int("app.port", 12345)
yiigo.Env.String("app.env", "dev")
```

#### Zipkin

```go
tracer, err := yiigo.NewZipkinTracer("http://localhost:9411/api/v2/spans",
    yiigo.WithZipkinTracerEndpoint("zipkin-test", "localhost"),
    yiigo.WithZipkinTracerSharedSpans(false),
    yiigo.WithZipkinTracerSamplerMod(1),
)

if err != nil {
    log.Fatal(err)
}

client, err := yiigo.NewZipkinClient(tracer)

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
yiigo.RegisterLogger(yiigo.AsDefault, "app.log")
yiigo.Logger.Info("hello world")

// other logger
yiigo.RegisterLogger("foo", "foo.log")
yiigo.UseLogger("foo").Info("hello world")
```

## Documentation

- [API Reference](https://godoc.org/github.com/iiinsomnia/yiigo)
- [TOML](https://github.com/toml-lang/toml)
- [Example](https://github.com/iiinsomnia/yiigo-example)

**Enjoy 😊**
