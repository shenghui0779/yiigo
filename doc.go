// Package yiigo 简单易用的 Golang 辅助库
//
// 使用 MySQL：
//
//    // default db
//    yiigo.RegisterDB("default", yiigo.MySQL, "root:root@tcp(localhost:3306)/test")
//
//    yiigo.DB.Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
//    // other db
//    yiigo.RegisterDB("foo", yiigo.MySQL, "root:root@tcp(localhost:3306)/test")
//
//    yiigo.UseDB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
// 使用 MongoDB：
//
//    // default mongodb
//    yiigo.RegisterMongoDB("default", "mongodb://username:password@localhost:27017")
//
//    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//    yiigo.Mongo.Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
//
//    // other mongodb
//    yiigo.RegisterMongoDB("foo", "mongodb://username:password@localhost:27017")
//
//    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//    yiigo.UseMongo("foo").Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
//
// 使用 Redis：
//
//    // default redis
//    yiigo.RegisterRedis("default", "localhost:6379")
//
//    conn, err := yiigo.Redis.Get()
//
//    if err != nil {
// 	      log.Fatal(err)
//    }
//
//    defer yiigo.Redis.Put(conn)
//
//    conn.Do("SET", "test_key", "hello world")
//
//    // other redis
//    yiigo.RegisterRedis("foo", "localhost:6379")
//
//    foo := yiigo.UseRedis("foo")
//    conn, err := foo.Get()
//
//    if err != nil {
// 	      log.Fatal(err)
//    }
//
//    defer foo.Put(conn)
//
//    conn.Do("SET", "test_key", "hello world")
//
// 使用配置文件：
//
//    // env.toml
//    //
//    // [app]
//    // env = "dev"
//    // debug = true
//    // port = 50001
//
//    yiigo.UseEnv("env.toml")
//
//    yiigo.Env.GetBool("app.debug", true)
//    yiigo.Env.GetInt("app.port", 12345)
//    yiigo.Env.GetString("app.env", "dev")
//
// 使用日志：
//
//    // default logger
//    yiigo.RegisterLogger("default", "app.log")
//    yiigo.Logger.Info("hello world")
//
//    // other logger
//    yiigo.RegisterLogger("foo", "foo.log")
//    yiigo.UseLogger("foo").Info("hello world")
//
// For more details, see the documentation for the types and methods.
//
package yiigo
