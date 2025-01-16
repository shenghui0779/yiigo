# yiigo tools

## 脚手架

```shell
go install github.com/shenghui0779/yiigo/cmd/yiigo@latest
```

#### 创建项目

分HTTP和gRPC两种，分别可创建单应用和多应用项目

##### 👉 HTTP

```shell
# 单应用
yiigo new demo
yiigo new demo --mod=xxx.yyy.com # 指定module名称
yiigo ent # 创建Ent默认实例
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── api
    │   ├── cmd
    │   ├── config
    │   ├── config.toml
    │   ├── main.go
    │   ├── middleware
    │   ├── router
    │   ├── service
    │   └── web
    ├── ent
    └── internal

# 多应用
yiigo new demo --apps=foo,bar
yiigo new demo --apps=foo --apps=bar
yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar
yiigo ent foo bar # 创建Ent实例
yiigo app hello # 创建应用
yiigo ent hello # 创建Ent实例
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── foo
    │   │   ├── api
    │   │   ├── cmd
    │   │   ├── config
    │   │   ├── config.toml
    │   │   ├── main.go
    │   │   ├── middleware
    │   │   ├── router
    │   │   ├── service
    │   │   └── web
    │   ├── bar
    │   └── hello
    ├── ent
    │   ├── foo
    │   ├── bar
    │   └── hello
    └── internal
```

##### 👉 gRPC

```shell
# 单应用
yiigo new demo --grpc
yiigo new demo --mod=xxx.yyy.com --grpc # 指定module名称
yiigo ent # 创建Ent默认实例
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── api
    │   │   ├── buf
    │   │   │   └── validate
    │   │   │       └── validate.proto
    │   │   ├── google
    │   │   │   └── api
    │   │   │       ├── annotations.proto
    │   │   │       └── http.proto
    │   │   └── greeter.proto
    │   ├── buf.gen.yaml
    │   ├── buf.yaml
    │   ├── cmd
    │   ├── config
    │   ├── config.toml
    │   ├── main.go
    │   ├── server
    │   └── service
    ├── ent
    └── internal

# 多应用
yiigo new demo --apps=foo,bar --grpc
yiigo new demo --apps=foo --apps=bar --grpc
yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar --grpc
yiigo ent foo bar # 创建Ent实例
yiigo app hello --grpc # 创建应用
yiigo ent hello # 创建Ent实例
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── foo
    │   │   ├── api
    │   │   │   ├── buf
    │   │   │   │   └── validate
    │   │   │   │       └── validate.proto
    │   │   │   ├── google
    │   │   │   │   └── api
    │   │   │   │       ├── annotations.proto
    │   │   │   │       └── http.proto
    │   │   │   └── greeter.proto
    │   │   ├── buf.gen.yaml
    │   │   ├── buf.yaml
    │   │   ├── cmd
    │   │   ├── config
    │   │   ├── config.toml
    │   │   ├── main.go
    │   │   ├── server
    │   │   └── service
    │   ├── bar
    │   └── hello
    ├── ent
    │   ├── foo
    │   ├── bar
    │   └── hello
    └── internal
```

## gg

为结构体生成`Get`方法，避免空指针导致Panic

Generate `Get` method for the structure to avoid panic caused by null pointer

#### 安装

```shell
go install github.com/shenghui0779/yiigo/cmd/gg@latest
```

#### 使用

```shell
# CLI
gg xxx.go

# go generate
//go:generate gg xxx.go
```
