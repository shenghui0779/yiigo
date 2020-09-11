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
- ORM使用 [gorm](https://gorm.io/)
- 日志使用 [zap](https://github.com/uber-go/zap)
- 包含一些实用的帮助方法，如：http、cypto、date、IP、SQL Builder 等

## Requirements

`Go1.11+`

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
    driver = "mysql"
    dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
    # dsn = "host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable" # pgsql
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

[log]

    [log.default]
    path = "app.log"
    max_size = 500
    max_age = 0
    max_backups = 0
    compress = true

[email]

    [email.default]
    host = "smtp.exmail.qq.com"
    port = 25
    username = ""
    password = ""

# apollo namespace

[apollo_test]
name = "yiigo"
```

- usage

```go
yiigo.Env("app.env").String("dev")
yiigo.Env("app.debug").Bool(true)
yiigo.Env("apollo_test.name").String("foo")
```

> ⚠️注意！
>
> 如果配置了 `apollo`，则：
>
> 1. `namespace` 默认包含 `application`；
> 2. `namespace` 中的配置项优先从 `apollo` 读取，若不存在，则从 `yiigo.toml` 中读取；
> 3. 若 `namespace` 不在 `apollo` 配置中，则其配置项从 `yiigo.toml` 中获取;

#### MySQL

```go
// default db
yiigo.DB().Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
yiigo.Orm().First(&User{}, 1)

// other db
yiigo.DB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
yiigo.Orm("foo").First(&User{}, 1)
```

#### MongoDB

```go
// default mongodb
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

defer cancel()

yiigo.Mongo().Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})

// other mongodb
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

defer cancel()

yiigo.Mongo("foo").Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
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
conn, err := yiigo.Redis("foo").Get()

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis("foo").Put(conn)

conn.Do("SET", "test_key", "hello world")
```

#### HTTP

```go
client, err := yiigo.NewHTTPClient(
    yiigo.WithHTTPMaxIdleConnsPerHost(1000),
    yiigo.WithHTTPMaxConnsPerHost(1000),
    yiigo.WithHTTPDefaultTimeout(time.Second*10),
)

if err != nil {
    log.Fatal(err)
}

b, err := client.Get("url...", yiigo.WithRequestTimeout(5*time.Second))

if err != nil {
    log.Fatal(err)
}

fmt.Println(string(b))
```

#### Logger

```go
// default logger
yiigo.Logger().Info("hello world")

// other logger
yiigo.Logger("foo").Info("hello world")
```

#### SQL Builder

- Query

```go
yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Where("id = ?", 1).ToQuery()
// SELECT * FROM user WHERE id = ?
// [1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Where("name = ? AND age > ?", "shenghui0779", 20).ToQuery()
// SELECT * FROM user WHERE name = ? AND age > ?
// [shenghui0779 20]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Select("id", "name", "age").Where("id = ?", 1).ToQuery()
// SELECT id, name, age FROM user WHERE id = ?
// [1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Distinct("name").Where("id = ?", 1).ToQuery()
// SELECT DISTINCT name FROM user WHERE id = ?
// [1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").LeftJoin("address", "user.id = address.user_id").Where("user.id = ?", 1).ToQuery()
// SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?
// [1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("address").Select("user_id", "COUNT(*) AS total").Group("user_id").Having("user_id = ?", 1).ToQuery()
// SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?
// [1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Where("age > ?", 20).Order("id DESC").Offset(5).Limit(10).ToQuery()
// SELECT * FROM user WHERE age > ? ORDER BY id DESC OFFSET 5 LIMIT 10
// [20]
```

- Insert

```go
yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").ToInsert(yiigo.X{
    "name": "shenghui0779",
    "age":  29,
})
// INSERT INTO user ( name, age ) VALUES ( ?, ? )
// [shenghui0779 29]
```

- Batch Insert

```go
yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").ToBatchInsert([]yiigo.X{
    {
        "name": "shenghui0779",
        "age":  29,
    },
    {
        "name": "iiinsomnia",
        "age":  30,
    },
})
// INSERT INTO user ( name, age ) VALUES ( ?, ? ), ( ?, ? )
// [shenghui0779 29 iiinsomnia 30]
```

- Update

```go
yiigo.NewSQLBuilder(yiigo.MySQL).Table("user").Where("id = ?", 1).ToUpdate(yiigo.X{
    "name": "shenghui0779",
    "age":  29,
})
// UPDATE user SET name = ?, age = ? WHERE id = ?
// [shenghui0779 29 1]

yiigo.NewSQLBuilder(yiigo.MySQL).Table("goods").Where("id = ?", 1).ToUpdate(yiigo.X{
    "amount": yiigo.Clause("amount * ? + ?", 2, 10),
    "price":  250,
})
// UPDATE goods SET amount = amount * ? + ?, price = ? WHERE id = ?
// [2 10 250 1]
```

- Delete

```go
yiigo.NewSQLBuilder(yiigo.MySQL).Where("id = ?", 1).ToDelete()
// DELETE FROM user WHERE id = ?
// [1]
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/shenghui0779/yiigo)
- [TOML](https://github.com/toml-lang/toml)
- [Example](https://github.com/shenghui0779/yiigo-example)

**Enjoy 😊**
