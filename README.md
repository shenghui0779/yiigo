# yiigo
Golang常用类库整合，用于开发API

### 获取：
go get github.com/iiinsomnia/yiigo

### 注：
##### [1]
 - 由于code.google.com被墙，导致一些托管在code.google.com上面的包go get不下来，可以到这里自行下载：
 - http://www.golangtc.com/download/package

 ##### [2]
 - 数据库默认没有主从，如果要使用主从库，如下：
 - 1、数据库操作文件用 mysqlx.go
 - 2、配置文件采用 dbx.ini 中的配置
