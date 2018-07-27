yiigo
===

简单易用的 Go Web 微框架

## 特点

- 支持多 [MySQL](https://github.com/jmoiron/sqlx) 连接
- 支持多 [mongo](http://labix.org/mgo) 连接
- 支持多 [redis](https://github.com/gomodule/redigo) 连接
- 采用 [zap](https://github.com/uber-go/zap) 日志记录
- 采用 [toml](https://github.com/pelletier/go-toml) 配置文件
- 采用 [glide](https://glide.sh) 管理依赖包
- 支持 [gomail](https://github.com/go-gomail/gomail) 邮件发送
- 支持 [session](http://www.gorillatoolkit.org/pkg/sessions) 存取
- 支持爬虫模拟登录

## 获取

```sh
# glide
glide init
glide get github.com/iiinsomnia/yiigo

# go get
go get github.com/iiinsomnia/yiigo
```

## 使用

#### 1、import yiigo

```go
package main

import "github.com/iiinsomnia/yiigo"

func main() {
    // 启用 mysql、mongo、redis
    err := yiigo.Bootstrap(true, true, true)

    if err != nil {
        yiigo.Logger.Panic(err.Error())
    }

    // coding...
}
```

#### 2、resolve dependencies

```sh
# 获取 yiigo 所需依赖包
glide update
```

## 文档

- [API Reference](https://godoc.org/github.com/IIInsomnia/yiigo)
- [Example](https://github.com/IIInsomnia/yiigo-example)

## 说明

- 在 `main.go` 所在目录创建 `env.toml` 配置文件，具体配置可以参考 `env.toml.example`
- `MySQL`、`mongo`、`redis` 多连接配置参考 `env.toml.example` 中的多数据库配置部分(注释部分)
- `golang.org` 上 `go get` 不下来的库，可以在这里[获取](https://github.com/golang)
- 如爬虫不需要模拟登录，则只需要使用 [goquery](https://github.com/PuerkitoBio/goquery) 即可

**Enjoy 😊**
