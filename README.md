# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GitHub release](https://img.shields.io/github/release/shenghui0779/yiigo.svg)](https://github.com/shenghui0779/yiigo/releases/latest)
[![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/shenghui0779/yiigo)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

Go 轻量级开发通用库

## Features

- 支持 [MySQL](https://github.com/go-sql-driver/mysql)
- 支持 [PostgreSQL](https://github.com/lib/pq)
- 支持 [SQLite3](https://github.com/mattn/go-sqlite3)
- 支持 [MongoDB](https://github.com/mongodb/mongo-go-driver)
- 支持 [Redis](https://github.com/gomodule/redigo)
- 支持 [NSQ](https://github.com/nsqio/go-nsq)
- 支持 [Apollo](https://github.com/philchia/agollo)
- 邮件使用 [gomail](https://github.com/go-gomail/gomail)
- 配置使用 [toml](https://github.com/pelletier/go-toml)
- SQL使用 [sqlx](https://github.com/jmoiron/sqlx)
- ORM推荐 [ent](https://github.com/facebook/ent)
- 日志使用 [zap](https://github.com/uber-go/zap)
- 包含一些实用的帮助方法，如：http、cypto、date、IP、SQL Builder 等

## Requirements

`Go1.15+`

## Installation

```sh
go get github.com/shenghui0779/yiigo
```

## Usage

#### Config

- `yiigo.toml`

```toml
[app]
env = "dev" # dev | beta | prod
debug = true

[apollo]
app_id = "test"
cluster = "default"
address = "127.0.0.1:8080"
namespace = ["apollo_test"]
cache_dir = "./"
accesskey_secret = ""
insecure_skip_verify = true

[db]

    [db.default]
    driver = "mysql" # mysql | postgres | sqlite3
    dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local" # mysql
    # dsn = "host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable" # postgres
    # dsn = "file::memory:?cache=shared" # sqlite3
    max_open_conns = 20
    max_idle_conns = 10
    conn_max_lifetime = 60 # 秒

[mongo]

    [mongo.default]
    dsn = "mongodb://username:password@localhost:27017"
    connect_timeout = 10 # 秒
    pool_size = 10
    max_conn_idle_time = 60 # 秒
    mode = "primary" # primary | primary_preferred | secondary | secondary_preferred | nearest

[redis]

    [redis.default]
    address = "127.0.0.1:6379"
    password = ""
    database = 0
    connect_timeout = 10 # 秒
    read_timeout = 10 # 秒
    write_timeout = 10 # 秒
    pool_size = 10
    pool_limit = 20
    idle_timeout = 60 # 秒
    wait_timeout = 10 # 秒
    prefill_parallelism = 0 # 预填充连接数

[nsq]
lookupd = ["127.0.0.1:4161"]
nsqd = "127.0.0.1:4150"

[email]

    [email.default]
    host = "smtp.exmail.qq.com"
    port = 25
    username = ""
    password = ""

[log]

    [log.default]
    path = "app.log"
    max_size = 500
    max_age = 0
    max_backups = 0
    compress = true

# apollo namespace

[apollo_namespace]
name = "yiigo"
```

- usage

```go
yiigo.Env("app.env").String("dev")
yiigo.Env("app.debug").Bool(true)
yiigo.Env("apollo_namespace.name").String("foo")
```

> ⚠️ 注意！
>
> 如果配置了 `apollo`，则：
>
> 1. `namespace` 默认包含 `application`；
> 2. `namespace` 中的配置项优先从 `apollo` 读取，若不存在，则从 `yiigo.toml` 中读取；
> 3. 若 `namespace` 不在 `apollo` 配置中，则其配置项从 `yiigo.toml` 中获取;

#### MySQL

```go
// default db
yiigo.DB().Get(&User{}, "SELECT * FROM user WHERE id = ?", 1)

// other db
yiigo.DB("other").Get(&User{}, "SELECT * FROM user WHERE id = ?", 1)
```

#### ORM(ent)

```go
import "<your_project>/ent"

// default driver
client := ent.NewClient(ent.Driver(yiigo.EntDriver()))

// other driver
client := ent.NewClient(ent.Driver(yiigo.EntDriver("other")))
```

#### MongoDB

```go
// default mongodb
yiigo.Mongo().Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})

// other mongodb
yiigo.Mongo("other").Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})
```

#### Redis

```go
// default redis
conn, err := yiigo.Redis().Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis().Put(conn)

conn.Do("SET", "test_key", "hello world")

// other redis
conn, err := yiigo.Redis("other").Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis("other").Put(conn)

conn.Do("SET", "test_key", "hello world")
```

#### HTTP

```go
// default client
yiigo.HTTPGet(context.Background(), "URL", yiigo.WithHTTPTimeout(5*time.Second))

// new client
client := yiigo.NewHTTPClient(*http.Client)
client.Get(context.Background(), "URL", yiigo.WithHTTPTimeout(5*time.Second))
```

#### Logger

```go
// default logger
yiigo.Logger().Info("hello world")

// other logger
yiigo.Logger("other").Info("hello world")
```

#### SQL Builder

> 😊 如果你不想手写SQL，可以使用 SQL Builder，用于 `yiigo.DB().Select()` 等；
>
> ⚠️ SQL Builder 作为辅助使用，目前支持的特性有限，复杂的SQL（如：子查询等）还需自己手写

```go
builder := yiigo.NewSQLBuilder(yiigo.MySQL)
```

- Query

```go
builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToQuery()
// SELECT * FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("name = ? AND age > ?", "shenghui0779", 20),
).ToQuery()
// SELECT * FROM user WHERE name = ? AND age > ?
// [shenghui0779 20]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.WhereIn("age IN (?)", []int{20, 30}),
).ToQuery()
// SELECT * FROM user WHERE age IN (?, ?)
// [20 30]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Select("id", "name", "age"),
    yiigo.Where("id = ?", 1),
).ToQuery()
// SELECT id, name, age FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Distinct("name"),
    yiigo.Where("id = ?", 1),
).ToQuery()
// SELECT DISTINCT name FROM user WHERE id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.LeftJoin("address", "user.id = address.user_id"),
    yiigo.Where("user.id = ?", 1),
).ToQuery()
// SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?
// [1]

builder.Wrap(
    yiigo.Table("address"),
    yiigo.Select("user_id", "COUNT(*) AS total"),
    yiigo.GroupBy("user_id"),
    yiigo.Having("user_id = ?", 1),
).ToQuery()
// SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?
// [1]

builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("age > ?", 20),
    yiigo.OrderBy("id DESC"),
    yiigo.Offset(5),
    yiigo.Limit(10),
).ToQuery()
// SELECT * FROM user WHERE age > ? ORDER BY id DESC OFFSET 5 LIMIT 10
// [20]

wrap1 := builder.Wrap(
	Table("user_1"),
	Where("id = ?", 2),
)

builder.Wrap(
    Table("user_0"),
    Where("id = ?", 1),
    Union(wrap1),
).ToQuery()
// SELECT * FROM user_0 WHERE id = ? UNION SELECT * FROM user_1 WHERE id = ?
// [1, 2]

builder.Wrap(
    Table("user_0"),
    Where("id = ?", 1),
    UnionAll(wrap1),
).ToQuery()
// SELECT * FROM user_0 WHERE id = ? UNION ALL SELECT * FROM user_1 WHERE id = ?
// [1, 2]
```

- Insert

```go
builder.Wrap(yiigo.Table("user")).ToInsert(yiigo.X{
    "name": "shenghui0779",
    "age":  29,
})
// INSERT INTO user (name, age) VALUES (?, ?)
// [shenghui0779 29]
```

- Batch Insert

```go
builder.Wrap(yiigo.Table("user")).ToBatchInsert([]yiigo.X{
    {
        "name": "shenghui0779",
        "age":  29,
    },
    {
        "name": "iiinsomnia",
        "age":  30,
    },
})
// INSERT INTO user (name, age) VALUES (?, ?), (?, ?)
// [shenghui0779 29 iiinsomnia 30]
```

- Update

```go
builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToUpdate(yiigo.X{
    "name": "shenghui0779",
    "age":  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [shenghui0779 29 1]

builder.Wrap(
    yiigo.Table("product"),
    yiigo.Where("id = ?", 1),
).ToUpdate(yiigo.X{
    "price": yiigo.Clause("price * ? + ?", 2, 100),
})
// UPDATE product SET price = price * ? + ? WHERE id = ?
// [2 100 1]
```

- Delete

```go
builder.Wrap(
    yiigo.Table("user"),
    yiigo.Where("id = ?", 1),
).ToDelete()
// DELETE FROM user WHERE id = ?
// [1]
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/shenghui0779/yiigo)
- [TOML](https://github.com/toml-lang/toml)
- [Example](https://github.com/shenghui0779/yiigo-example)

**Enjoy 😊**
