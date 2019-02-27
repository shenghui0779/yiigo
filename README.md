# yiigo

[![GoDoc](https://godoc.org/github.com/IIInsomnia/yiigo?status.svg)](https://godoc.org/github.com/IIInsomnia/yiigo)
[![GitHub release](https://img.shields.io/github/release/IIInsomnia/yiigo.svg)](https://github.com/IIInsomnia/yiigo/releases/latest)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

简单易用的 Golang 辅助库，让 Golang 开发更简单

## 特点

- 支持多 [MySQL](https://github.com/go-sql-driver/mysql) 连接
- 支持多 [PostgreSQL](https://github.com/lib/pq) 连接
- 支持多 [mongo](https://github.com/mongodb/mongo-go-driver) 连接
- 支持多 [redis](https://github.com/gomodule/redigo) 连接
- 支持 [gomail](https://github.com/go-gomail/gomail) 邮件发送
- 支持 [toml](https://github.com/pelletier/go-toml) 文件配置
- 采用 [sqlx](https://github.com/jmoiron/sqlx) 处理SQL查询
- 采用 [zap](https://github.com/uber-go/zap) 日志记录

## 获取

```sh
go get github.com/iiinsomnia/yiigo
```

## 使用

#### MySQL

```go
// default db
yiigo.RegisterDB("default", yiigo.MySQL, "root:root@tcp(localhost:3306)/test")

yiigo.DB.Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)

// other db
yiigo.RegisterDB("foo", yiigo.MySQL, "root:root@tcp(localhost:3306)/test")

yiigo.UseDB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
```

#### MongoDB

```go
// default mongodb
yiigo.RegisterMongoDB("default", "mongodb://username:password@localhost:27017")

ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
yiigo.Mongo.Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})

// other mongodb
yiigo.RegisterMongoDB("foo", "mongodb://username:password@localhost:27017")

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

## 文档

- [API Reference](https://godoc.org/github.com/IIInsomnia/yiigo)
- [Example](https://github.com/IIInsomnia/yiigo-example)

## 说明

- 支持 Go1.11+
- 配置文件相关语法参考 [toml](https://github.com/toml-lang/toml)
- 爬虫可能需要的两个库：
    1. 页面 DOM 处理：[goquery](https://github.com/PuerkitoBio/goquery)
    2. GBK 转 UTF8：[iconv](https://github.com/qiniu/iconv)

**Enjoy 😊**
