# yiigo
Golang集成常用类库并封装，用于WEB、API和爬虫开发

## 特点

* 支持多数据库连接
* 支持 `redis`
* 支持 `mongo`
* 支持爬虫模拟登录
* 支持邮件发送
* 支持日志记录
* 采用 `ini` 配置文件

## 获取

```go
go get github.com/iiinsomnia/yiigo
```

## 使用

```go
package main

import "github.com/iiinsomnia/yiigo"

func main() {
    b := yiigo.New()

	b.EnableMongo() // 启用mongo
	b.EnableRedis() // 启用redis
	b.Bootstrap()
}
```

```go
// 设置配置文件路径，默认为env.ini，具体配置参考env.ini.example
yiigo.SetEnv("myenv.ini")
```

## 说明
* 在 `main.go` 所在目录创建 `env.ini` 配置文件，具体配置可以参考 `env.ini.example`
* 目前数据库仅针对MySQL封装，多数据库连接只需在`ini`文件中配置多个即可
* code.google.com 上 go get 不下来的库，可以在这里[获取](https://github.com/golang)
* 如爬虫不需要模拟登录，则只需要使用 [goquery](https://github.com/PuerkitoBio/goquery) 即可
* 具体使用可以参考 [yiigo-example](https://github.com/IIInsomnia/yiigo-example)
* [DB文档](http://jmoiron.github.io/sqlx/)
* [Mongo文档](http://labix.org/mgo)
* [Redis文档](http://godoc.org/github.com/garyburd/redigo/redis)