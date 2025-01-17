# 脚手架

自动生成项目，支持 `HTTP` 和 `gRPC`，并同时支持创建「单应用」和「多应用」

## 安装

```shell
go install github.com/shenghui0779/yiigo/cmd/yiigo@latest
```

## 创建项目

#### 单应用

```shell
# http
yiigo new demo
yiigo new demo --mod=xxx.com/demo # 指定module名称
.
├── go.mod
├── go.sum
├── Dockerfile
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
    └── internal

# grpc
yiigo new demo --grpc
yiigo new demo --mod=xxx.com/demo --grpc # 指定module名称
.
├── go.mod
├── go.sum
├── Dockerfile
└── pkg
    ├── app
    │   ├── api
    │   │   └── greeter.proto
    │   ├── buf.gen.yaml
    │   ├── buf.lock
    │   ├── buf.yaml
    │   ├── cmd
    │   ├── config
    │   ├── config.toml
    │   ├── main.go
    │   ├── server
    │   └── service
    └── internal
```

#### 多应用

```shell
# http
yiigo new demo --apps=foo,bar
yiigo new demo --mod=xxx.com/demo --apps=foo,bar
yiigo new demo --mod=xxx.com/demo --apps=foo --apps=bar
.
├── go.mod
├── go.sum
├── foo.dockerfile
├── bar.dockerfile
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
    │   │   ├── ...
    │   │   └── main.go
    └── internal

# grpc
yiigo new demo --apps=foo,bar --grpc
yiigo new demo --mod=xxx.com/demo --apps=foo,bar --grpc
yiigo new demo --mod=xxx.com/demo --apps=foo --apps=bar --grpc
.
├── go.mod
├── go.sum
├── foo.dockerfile
├── bar.dockerfile
└── pkg
    ├── app
    │   ├── foo
    │   │   ├── api
    │   │   │   └── greeter.proto
    │   │   ├── buf.gen.yaml
    │   │   ├── buf.lock
    │   │   ├── buf.yaml
    │   │   ├── cmd
    │   │   ├── config
    │   │   ├── config.toml
    │   │   ├── main.go
    │   │   ├── server
    │   │   └── service
    │   ├── bar
    │   │   ├── ...
    │   │   └── main.go
    └── internal
```

## 创建应用

```shell
# 多应用项目适用，需在项目根目录执行（即：go.mod所在目录）
yiigo app foo # 创建应用 -- foo
yiigo app foo bar # 创建两个应用 -- foo 和 bar
.
├── go.mod
├── go.sum
├── foo.dockerfile
├── bar.dockerfile
└── pkg
    ├── app
    │   ├── foo
    │   │   ├── ...
    │   │   └── main.go
    │   ├── bar
    │   │   ├── ...
    │   │   └── main.go
    └── internal
```

## 创建Ent实例

#### 单实例

```shell
yiigo ent
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    ├── ent
    │   ├── ...
    │   └── schema
    └── internal
```

#### 多实例

```shell
# 创建Ent实例 -- foo 和 bar
yiigo ent foo bar
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    ├── ent
    │   ├── foo
    │   │   ├── ...
    │   │   └── schema
    │   ├── bar
    │   │   ├── ...
    │   │   └── schema
    └── internal
```
