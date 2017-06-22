# yiigo
Golang集成常用类库并封装，用于API开发和爬虫，支持多数据库连接

## 获取：

```go
go get github.com/iiinsomnia/yiigo
```

## 使用：

```go
package main

import "github.com/iiinsomnia/yiigo"

func main() {
    b := yiigo.New()
	b.EnableMongo() // 启用mongo
	b.EnableRedis() // 启用redis
	b.Run()
}
```

```go
// 设置多个连接配置，默认配置为mysql
yiigo.SetMySQL("mysql1", "mysql2", "mysql3")
```

```
注意：配置和日志一定要优先加载初始化，其余按需初始化即可
```

## 说明：
* 目前数据库仅针对MySQL封装
* 由于 code.google.com 被墙，导致一些托管在 code.google.com 上面的包 go get 不下来，可以到这里自行下载：[下载和使用](http://www.golangtc.com/download/package)
* 如果爬虫不需要模拟登录，则只需要使用 [goquery](https://github.com/PuerkitoBio/goquery) 即可
* API开发的具体使用方法可以[参考这里](https://github.com/IIInsomnia/yiigo-example)