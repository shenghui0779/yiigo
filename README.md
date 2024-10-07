# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org) [![GitHub release](https://img.shields.io/github/release/shenghui0779/yiigo.svg)](https://github.com/shenghui0779/yiigo/releases/latest) [![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/shenghui0779/yiigo) [![Apache 2.0 license](http://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](http://opensource.org/licenses/apache2.0)

一个好用的Go项目脚手架和工具包

## 脚手架

```shell
# go1.23+
go install github.com/shenghui0779/yiigo/cmd/yiigo@latest
```

#### 创建项目

分HTTP和gRPC两种，分别可创建单应用和多应用项目

##### 👉 HTTP

```shell
# 单应用
yiigo new demo
yiigo new demo --mod=xxx.yyy.com # 指定module名称
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── Dockerfile
    │   ├── api
    │   │   ├── controller
    │   │   ├── middleware
    │   │   ├── router
    │   │   └── service
    │   ├── cmd
    │   ├── config.toml
    │   ├── ent
    │   ├── main.go
    │   └── web
    └── lib

# 多应用
yiigo new demo --apps=foo,bar
yiigo new demo --apps=foo --apps=bar
yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar
yiigo app hello # 新增应用
.
├── go.mod
├── go.sum
└── pkg
    ├── app
    │   ├── foo
    │   │   ├── api
    │   │   │   ├── controller
    │   │   │   ├── middleware
    │   │   │   ├── router
    │   │   │   └── service
    │   │   ├── cmd
    │   │   ├── config.toml
    │   │   ├── ent
    │   │   ├── main.go
    │   │   └── web
    │   ├── bar
    │   └── hello
    └── lib
```

##### 👉 gRPC

```shell
# 单应用
yiigo new demo --grpc
yiigo new demo --mod=xxx.yyy.com --grpc # 指定module名称
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
    │   ├── config.toml
    │   ├── ent
    │   ├── main.go
    │   ├── server
    │   └── service
    └── lib

# 多应用
yiigo new demo --apps=foo,bar --grpc
yiigo new demo --apps=foo --apps=bar --grpc
yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar --grpc
yiigo app hello --grpc # 新增应用
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
    │   │   ├── config.toml
    │   │   ├── ent
    │   │   ├── main.go
    │   │   ├── server
    │   │   └── service
    │   ├── bar
    │   └── hello
    └── lib
```

## 工具包

```shell
# go1.20+
go get -u github.com/shenghui0779/yiigo
```

### Features

- Hash - 封装便于使用
- Http - 封装便于使用(支持文件上传)
- Crypto - 封装便于使用(支持 AES & RSA)
- Validator - 支持汉化和自定义规则
- 基于 Redis 的分布式锁
- 基于 sqlx 的轻量SQLBuilder
- 基于泛型的无限菜单分类层级树
- errgroup - 基于官方版本改良，支持并发协程数量控制
- Value - 用于处理 `k-v` 格式化的场景，如：生成签名串 等
- Location - 距离、方位角、经纬度与平面直角坐标系的相互转化
- TimeWheel - 简单实用的单层时间轮(支持一次性和多次重试任务)
- 实用的辅助方法：IP、file、time、slice、string、version compare 等

#### SQL Builder

> ⚠️ 目前支持的特性有限，复杂的SQL（如：子查询等）还需自己手写

```go
builder := yiigo.NewSQLBuilder(*sqlx.DB, func(ctx context.Context, query string, args ...any) {
    fmt.Println(query, args)
})
```

##### 👉 Query

```go
ctx := context.Background()

type User struct {
    ID     int    `db:"id"`
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

var record User
builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).One(ctx, &record)
// SELECT * FROM user WHERE id = ?
// [1]

var records []User
builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("name = ? AND age > ?", "shenghui0779", 20),
).All(ctx, &records)
// SELECT * FROM user WHERE name = ? AND age > ?
// [shenghui0779 20]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.WhereIn("age IN (?)", []int{20, 30}),
).All(...)
// SELECT * FROM user WHERE age IN (?, ?)
// [20 30]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Select("id", "name", "age"),
    yiigo.Where("id = ?", 1),
).One(...)
// SELECT id, name, age FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Distinct("name"),
    yiigo.Where("id = ?", 1),
).One(...)
// SELECT DISTINCT name FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.LeftJoin("address", "user.id = address.user_id"),
    yiigo.Where("user.id = ?", 1),
).One(...)
// SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?
// [1]

builder.Wrap(
    yiigo.Table("address"),
    yiigo.Select("user_id", "COUNT(*) AS total"),
    yiigo.GroupBy("user_id"),
    yiigo.Having("user_id = ?", 1),
).All(...)
// SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("age > ?", 20),
    yiigo.OrderBy("age ASC", "id DESC"),
    yiigo.Offset(5),
    yiigo.Limit(10),
).All(...)
// SELECT * FROM user WHERE age > ? ORDER BY age ASC, id DESC LIMIT ? OFFSET ?
// [20, 10, 5]

wrap1 := builder.Wrap(
    yiigo.Table("user_1"),
    yiigo.Where("id = ?", 2),
)

builder.Wrap(
    yiigo.Table("user_0"),
    yiigo.Where("id = ?", 1),
    yiigo.Union(wrap1),
).All(...)
// (SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?)
// [1, 2]

builder.Wrap(
    yiigo.Table("user_0"),
    yiigo.Where("id = ?", 1),
    yiigo.UnionAll(wrap1),
).All(...)
// (SELECT * FROM user_0 WHERE id = ?) UNION ALL (SELECT * FROM user_1 WHERE id = ?)
// [1, 2]

builder.Wrap(
    yiigo.Table("user_0"),
    yiigo.WhereIn("age IN (?)", []int{10, 20}),
    yiigo.Limit(5),
    yiigo.Union(
        builder.Wrap(
            yiigo.Table("user_1"),
            yiigo.Where("age IN (?)", []int{30, 40}),
            yiigo.Limit(5),
        ),
    ),
).All(...)
// (SELECT * FROM user_0 WHERE age IN (?, ?) LIMIT ?) UNION (SELECT * FROM user_1 WHERE age IN (?, ?) LIMIT ?)
// [10, 20, 5, 30, 40, 5]
```

##### 👉 Insert

```go
ctx := context.Background()

type User struct {
    ID     int64  `db:"-"`
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(Table("user")).Insert(ctx, &User{
    Name: "yiigo",
    Age:  29,
})
// INSERT INTO user (name, age) VALUES (?, ?)
// [yiigo 29]

builder.Wrap(yiigo.Table("user")).Insert(ctx, yiigo.X{
    "name": "yiigo",
    "age":  29,
})
// INSERT INTO user (name, age) VALUES (?, ?)
// [yiigo 29]
```

##### 👉 Batch Insert

```go
ctx := context.Background()

type User struct {
    ID     int64  `db:"-"`
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(Table("user")).BatchInsert(ctx, []*User{
    {
        Name: "shenghui0779",
        Age:  20,
    },
    {
        Name: "yiigo",
        Age:  29,
    },
})
// INSERT INTO user (name, age) VALUES (?, ?), (?, ?)
// [shenghui0779 20 yiigo 29]

builder.Wrap(yiigo.Table("user")).BatchInsert(ctx, []yiigo.X{
    {
        "name": "shenghui0779",
        "age":  20,
    },
    {
        "name": "yiigo",
        "age":  29,
    },
})
// INSERT INTO user (name, age) VALUES (?, ?), (?, ?)
// [shenghui0779 20 yiigo 29]
```

##### 👉 Update

```go
ctx := context.Background()

type User struct {
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).Update(ctx, &User{
    Name: "yiigo",
    Age:  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [yiigo 29 1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).Update(ctx, yiigo.X{
    "name": "yiigo",
    "age":  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [yiigo 29 1]

builder.Wrap(
    yiigo.Table("product"),
    yiigo.Where("id = ?", 1),
).Update(ctx, yiigo.X{
    "price": yiigo.SQLExpr("price * ? + ?", 2, 100),
})
// UPDATE product SET price = price * ? + ? WHERE id = ?
// [2 100 1]
```

##### 👉 Delete

```go
ctx := context.Background()

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).Delete(ctx)
// DELETE FROM user WHERE id = ?
// [1]

builder.Wrap(yiigo.Table("user")).Truncate(ctx)
// TRUNCATE user
```

##### 👉 Transaction

```go
builder.Transaction(context.Background(), func(ctx context.Context, tx yiigo.TXBuilder) error {
    _, err := tx.Wrap(
        yiigo.Table("address"),
        yiigo.Where("user_id = ?", 1),
    ).Update(ctx, yiigo.X{"default": 0})
    if err != nil {
        return err
    }

    _, err = tx.Wrap(
        yiigo.Table("address"),
        yiigo.Where("id = ?", 1),
    ).Update(ctx, yiigo.X{"default": 1})

    return err
})
```

**Enjoy 😊**
