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
// 修改配置文件，默认为env.ini，具体配置参考env.ini.example
yiigo.SetEnv("myenv.ini")

// 修改日志配置文件，默认为log.xml，具体配置参考log.xml.example
yiigo.SetLog("mylog.xml")
```

## 说明：
* 目前数据库仅针对MySQL封装，多数据库连接只需在`ini`文件中配置多个即可
* code.google.com 上 go get 不下来的库，可以在这里[获取](https://github.com/golang)
* 如爬虫不需要模拟登录，则只需要使用 [goquery](https://github.com/PuerkitoBio/goquery) 即可
* API开发的具体使用方法可以[参考这里](https://github.com/IIInsomnia/yiigo-example)