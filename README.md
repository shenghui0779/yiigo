# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org) [![GitHub release](https://img.shields.io/github/release/shenghui0779/yiigo.svg)](https://github.com/shenghui0779/yiigo/releases/latest) [![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/shenghui0779/yiigo) [![Apache 2.0 license](http://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](http://opensource.org/licenses/apache2.0)

一个好用的 Go 开发工具库

```sh
go get -u github.com/shenghui0779/yiigo
```

## Features

- NSQ
- Hash
- HTTP
- Crypto
- Validator
- 轻量的用于 `sqlx` 的 SQL Builder
- 基于 Redis 的简单分布式锁
- Websocket 简单使用封装
  - Dialer - 读写失败支持重连
  - Upgrader - 支持授权校验
- 简单实用的单时间轮（支持一次性和多次重试任务）
- 用于处理 `k-v` 需要格式化的场景(如：生成签名串)的value包
- 实用的辅助方法，包含：IP、file、time、slice、string、version compare 等

#### SQL Builder

> 😊 为不想手写SQL的你生成SQL语句，用于 `sqlx` 的相关方法；
>
> ⚠️ 作为辅助方法，目前支持的特性有限，复杂的SQL（如：子查询等）还需自己手写

```go
builder := yiigo.NewMySQLBuilder()
// builder := yiigo.NewSQLBuilder(yiigo.MySQL)
```

- Query

```go
ctx := context.Background()

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToQuery(ctx)
// SELECT * FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("name = ? AND age > ?", "shenghui0779", 20),
).ToQuery(ctx)
// SELECT * FROM user WHERE name = ? AND age > ?
// [shenghui0779 20]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.WhereIn("age IN (?)", []int{20, 30}),
).ToQuery(ctx)
// SELECT * FROM user WHERE age IN (?, ?)
// [20 30]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Select("id", "name", "age"),
    yiigo.Where("id = ?", 1),
).ToQuery(ctx)
// SELECT id, name, age FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Distinct("name"),
    yiigo.Where("id = ?", 1),
).ToQuery(ctx)
// SELECT DISTINCT name FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.LeftJoin("address", "user.id = address.user_id"),
    yiigo.Where("user.id = ?", 1),
).ToQuery(ctx)
// SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?
// [1]

builder.Wrap(
    yiigo.Table("address"),
    yiigo.Select("user_id", "COUNT(*) AS total"),
    yiigo.GroupBy("user_id"),
    yiigo.Having("user_id = ?", 1),
).ToQuery(ctx)
// SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("age > ?", 20),
    yiigo.OrderBy("age ASC", "id DESC"),
    yiigo.Offset(5),
    yiigo.Limit(10),
).ToQuery(ctx)
// SELECT * FROM user WHERE age > ? ORDER BY age ASC, id DESC LIMIT ? OFFSET ?
// [20, 10, 5]

wrap1 := builder.Wrap(
    Table("user_1"),
    Where("id = ?", 2),
)

builder.Wrap(
    Table("user_0"),
    Where("id = ?", 1),
    Union(wrap1),
).ToQuery(ctx)
// (SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?)
// [1, 2]

builder.Wrap(
    Table("user_0"),
    Where("id = ?", 1),
    UnionAll(wrap1),
).ToQuery(ctx)
// (SELECT * FROM user_0 WHERE id = ?) UNION ALL (SELECT * FROM user_1 WHERE id = ?)
// [1, 2]

builder.Wrap(
    Table("user_0"),
    WhereIn("age IN (?)", []int{10, 20}),
    Limit(5),
    Union(
        builder.Wrap(
            Table("user_1"),
            Where("age IN (?)", []int{30, 40}),
            Limit(5),
        ),
    ),
).ToQuery(ctx)
// (SELECT * FROM user_0 WHERE age IN (?, ?) LIMIT ?) UNION (SELECT * FROM user_1 WHERE age IN (?, ?) LIMIT ?)
// [10, 20, 5, 30, 40, 5]
```

- Insert

```go
ctx := context.Background()

type User struct {
    ID     int    `db:"-"`
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(Table("user")).ToInsert(ctx, &User{
    Name: "yiigo",
    Age:  29,
})
// INSERT INTO user (name, age) VALUES (?, ?)
// [yiigo 29]

builder.Wrap(yiigo.Table("user")).ToInsert(ctx, yiigo.X{
    "name": "yiigo",
    "age":  29,
})
// INSERT INTO user (name, age) VALUES (?, ?)
// [yiigo 29]
```

- Batch Insert

```go
ctx := context.Background()

type User struct {
    ID     int    `db:"-"`
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(Table("user")).ToBatchInsert(ctx, []*User{
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

builder.Wrap(yiigo.Table("user")).ToBatchInsert(ctx, []yiigo.X{
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

- Update

```go
ctx := context.Background()

type User struct {
    Name   string `db:"name"`
    Age    int    `db:"age"`
    Phone  string `db:"phone,omitempty"`
}

builder.Wrap(
    Table("user"),
    Where("id = ?", 1),
).ToUpdate(ctx, &User{
    Name: "yiigo",
    Age:  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [yiigo 29 1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToUpdate(ctx, yiigo.X{
    "name": "yiigo",
    "age":  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [yiigo 29 1]

builder.Wrap(
    yiigo.Table("product"),
    yiigo.Where("id = ?", 1),
).ToUpdate(ctx, yiigo.X{
    "price": yiigo.SQLExpr("price * ? + ?", 2, 100),
})
// UPDATE product SET price = price * ? + ? WHERE id = ?
// [2 100 1]
```

- Delete

```go
ctx := context.Background()

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToDelete(ctx)
// DELETE FROM user WHERE id = ?
// [1]

builder.Wrap(Table("user")).ToTruncate(ctx)
// TRUNCATE user
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/shenghui0779/yiigo)
- [api-tpl-go](https://github.com/shenghui0779/api-tpl-go)
- [api-tpl-rs](https://github.com/shenghui0779/api-tpl-rs)

**Enjoy 😊**
