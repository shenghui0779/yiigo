# yiigo

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org) [![GitHub release](https://img.shields.io/github/release/shenghui0779/yiigo.svg)](https://github.com/shenghui0779/yiigo/releases/latest) [![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/shenghui0779/yiigo) [![Apache 2.0 license](http://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](http://opensource.org/licenses/apache2.0)

一个好用的轻量级 Go 开发通用库

## Features

- 支持 [MySQL](https://github.com/go-sql-driver/mysql)
- 支持 [PostgreSQL](https://github.com/lib/pq)
- 支持 [SQLite3](https://github.com/mattn/go-sqlite3)
- 支持 [MongoDB](https://github.com/mongodb/mongo-go-driver)
- 支持 [Redis](https://github.com/gomodule/redigo)
- 支持 [NSQ](https://github.com/nsqio/go-nsq)
- SQL使用 [sqlx](https://github.com/jmoiron/sqlx)
- ORM推荐 [ent](https://github.com/ent/ent)
- 日志使用 [zap](https://github.com/uber-go/zap)
- gRPC Client 连接池
- 轻量的 SQL Builder
- 环境配置使用 [dotenv](https://github.com/joho/godotenv)，支持热加载
- 实用的辅助方法，包含：http、cypto、date、IP、validator、version compare 等

## Installation

```sh
go get -u github.com/shenghui0779/yiigo
```

## Usage

#### ENV

- load

```go
// 默认加载当前目录下的`.env`文件
yiigo.LoadEnv()

// 加载指定配置文件
yiigo.LoadEnv(yiigo.WithEnvFile("mycfg.env"))

// 热加载
yiigo.LoadEnv(yiigo.WithEnvWatcher(onchanges...))
```

- `.env`

```sh
ENV=dev
```

- usage

```go
fmt.Println(os.Getenv("ENV"))
// output: dev
```

#### DB

- register

```go
yiigo.Init(
    yiigo.WithMySQL(yiigo.Default, &yiigo.DBConfig{
        DSN: "dsn",
        Options: &yiigo.DBOptions{
            MaxOpenConns:    20,
            MaxIdleConns:    10,
            ConnMaxLifetime: 10 * time.Minute,
            ConnMaxIdleTime: 5 * time.Minute,
        },
    }),

    yiigo.WithMySQL("other", &yiigo.DBConfig{
        DSN: "dsn",
        Options: &yiigo.DBOptions{
            MaxOpenConns:    20,
            MaxIdleConns:    10,
            ConnMaxLifetime: 10 * time.Minute,
            ConnMaxIdleTime: 5 * time.Minute,
        },
    }),
)
```

- sqlx

```go
// default db
yiigo.DB().Get(&User{}, "SELECT * FROM user WHERE id = ?", 1)

// other db
yiigo.DB("other").Get(&User{}, "SELECT * FROM user WHERE id = ?", 1)
```

- ent

```go
import "<your_project>/ent"

// default driver
client := ent.NewClient(ent.Driver(yiigo.EntDriver()))

// other driver
client := ent.NewClient(ent.Driver(yiigo.EntDriver("other")))
```

#### MongoDB

```go
// register
yiigo.Init(
    yiigo.WithMongo(yiigo.Default, "dsn"),
    yiigo.WithMongo("other", "dsn"),
)

// default mongodb
yiigo.Mongo().Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})

// other mongodb
yiigo.Mongo("other").Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})
```

#### Redis

```go
// register
yiigo.Init(
    yiigo.WithRedis(yiigo.Default, &yiigo.RedisConfig{
        Addr: "addr",
        Options: &yiigo.RedisOptions{
            ConnTimeout:  10 * time.Second,
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            PoolSize:     10,
            IdleTimeout:  5 * time.Minute,
        },
    }),

    yiigo.WithRedis("other", &yiigo.RedisConfig{
        Addr: "addr",
        Options: &yiigo.RedisOptions{
            ConnTimeout:  10 * time.Second,
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            PoolSize:     10,
            IdleTimeout:  5 * time.Minute,
        },
    }),
)

// default redis
conn, err := yiigo.Redis().Get(context.Background())

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis().Put(conn)

conn.Do("SET", "test_key", "hello world")

// other redis
conn, err := yiigo.Redis("other").Get(context.Background())

if err != nil {
    log.Fatal(err)
}

defer yiigo.Redis("other").Put(conn)

conn.Do("SET", "test_key", "hello world")
```

#### Logger

```go
// register
yiigo.Init(
    yiigo.WithLogger(yiigo.Default, yiigo.LoggerConfig{
        Filename: "filename",
        Options: &yiigo.LoggerOptions{
            Stderr: true,
        },
    }),

    yiigo.WithLogger("other", yiigo.LoggerConfig{
        Filename: "filename",
        Options: &yiigo.LoggerOptions{
            Stderr: true,
        },
    }),
)

// default logger
yiigo.Logger().Info("hello world")

// other logger
yiigo.Logger("other").Info("hello world")
```

#### gRPC Pool

```go
// create pool
pool := yiigo.NewGrpcPool(&yiigo.GrpcPoolConfig{
    Dialer: func() (*grpc.ClientConn, error) {
        return grpc.DialContext(context.Background(), "target",
            grpc.WithInsecure(),
            grpc.WithBlock(),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:    time.Second * 30,
                Timeout: time.Second * 10,
            }),
        )
    },
    Options: &yiigo.PoolOptions{
        PoolSize:    10,
        IdleTimeout: 5 * time.Minute,
    },
})

// use pool
conn, err := pool.Get(context.Background())

if err != nil {
    return err
}

defer pool.Put(conn)

// coding...
```

#### HTTP

```go
// default client
ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
yiigo.HTTPGet(ctx, "URL")

// new client
client := yiigo.NewHTTPClient(*http.Client)

ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
client.Do(ctx, http.MethodGet, "URL", nil)
```

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
    "price": yiigo.Clause("price * ? + ?", 2, 100),
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
- [Example](https://github.com/shenghui0779/yiigo-example)

**Enjoy 😊**
