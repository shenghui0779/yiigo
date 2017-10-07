# yiigo
Golang常用优秀库封装，用于API、WEB和爬虫开发

## 特点

* 支持 `MySQL` 多数据库连接
* 支持 `MySQL` 主从分离
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
    //启用 mysql、mongo、redis
    err := yiigo.Bootstrap(true, true, true)

    if err != nil {
        panic(err)
    }
}
```

## 说明
* 在 `main.go` 所在目录创建 `env.ini` ENV配置文件，具体配置可以参考 `env.ini.dev`
* 在 `main.go` 所在目录创建 `log.xml` 日志配置文件，具体配置可以参考 `log.xml.dev` 和 `log.xml.prod`
* `MySQL`多数据库和主从分离配置参考`env.ini.dev`中d的多数据库配置部分(注释部分)
* code.google.com 上 `go get` 不下来的库，可以在这里[获取](https://github.com/golang)
* 如爬虫不需要模拟登录，则只需要使用 [goquery](https://github.com/PuerkitoBio/goquery) 即可
* 具体使用可以参考 [yiigo-example](https://github.com/IIInsomnia/yiigo-example)