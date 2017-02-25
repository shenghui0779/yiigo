# yiigo

Golang常用类库整合，用于API开发和爬虫，支持多数据库连接

### 获取：

```go
go get github.com/iiinsomnia/yiigo
```

### 使用：

```go
package main

import "github.com/iiinsomnia/yiigo"

func main() {
    yiigo.LoadEnvConfig() // 加载配置
    yiigo.InitLogger()    // 初始化日志
    yiigo.InitDB()        // 初始化DB，没有指定配置名称，则默认为："db"
    yiigo.InitMongo()     // 初始化MongoDB
    yiigo.InitRedis()     // 初始化Redis
}
```

```go
yiigo.InitDB("db1", "db2", "db3") // 初始化多个DB
```

```
注意：配置和日志一定要优先加载初始化，其余按需初始化即可
```

### 说明：
* 由于 code.google.com 被墙，导致一些托管在 code.google.com 上面的包 go get 不下来，可以到这里自行下载：[下载和使用](http://www.golangtc.com/download/package)
* 具体使用方法可以[参考这里](https://github.com/IIInsomnia/yiigo-example)