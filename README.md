# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/iiinsomnia/yiigo?status.svg)](https://godoc.org/github.com/iiinsomnia/yiigo)
[![GitHub release](https://img.shields.io/github/release/IIInsomnia/yiigo.svg)](https://github.com/iiinsomnia/yiigo/releases/latest)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

a simple and light Golang library which makes Golang development easier!

## Features

- Support [MySQL](https://github.com/go-sql-driver/mysql)
- Support [PostgreSQL](https://github.com/lib/pq)
- Support [MongoDB](https://github.com/mongodb/mongo-go-driver)
- Support [Redis](https://github.com/gomodule/redigo)
- Use [gomail](https://github.com/go-gomail/gomail) for email sending
- Use [toml](https://github.com/pelletier/go-toml) for configuration
- Use [sqlx](https://github.com/jmoiron/sqlx) for SQL executing
- Use [zap](https://github.com/uber-go/zap) for logging
## Requirements

Go1.11+

## Installation

```sh
go get github.com/iiinsomnia/yiigo
```

## Usage

#### MySQL

```go
// default db
yiigo.RegisterDB("default", yiigo.MySQL, "root:root@tcp(localhost:3306)/test")

yiigo.DB.Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)

// other db
yiigo.RegisterDB("foo", yiigo.MySQL, "root:root@tcp(localhost:3306)/foo")

yiigo.UseDB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
```

#### MongoDB

```go
// default mongodb
yiigo.RegisterMongoDB("default", "mongodb://localhost:27017")

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
yiigo.RegisterRedis("default", "localhost:6379")

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

yiigo.Env.GetBool("app.debug", true)
yiigo.Env.GetInt("app.port", 12345)
yiigo.Env.GetString("app.env", "dev")
```

#### Logger

```go
// default logger
yiigo.RegisterLogger("default", "app.log")
yiigo.Logger.Info("hello world")

// other logger
yiigo.RegisterLogger("foo", "foo.log")
yiigo.UseLogger("foo").Info("hello world")
```

## Documentation

- [API Reference](https://godoc.org/github.com/iiinsomnia/yiigo)
- [toml](https://github.com/toml-lang/toml)
- [Example](https://github.com/iiinsomnia/yiigo-example)

**Enjoy 😊**
