// Package yiigo makes Golang development easier !
//
// Basic usage
//
// MySQL
//
//    // default db
//    yiigo.DB().Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
//    // other db
//    yiigo.DB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
// MongoDB
//
//    // default mongodb
//    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//
//    defer cancel()
//
//    yiigo.Mongo().Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
//
//    // other mongodb
//    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//
//    defer cancel()
//
//    yiigo.Mongo("foo").Database("test").Collection("numbers").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
//
// Redis
//
//    // default redis
//    conn, err := yiigo.Redis().Get()
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    defer yiigo.Redis().Put(conn)
//
//    conn.Do("SET", "test_key", "hello world")
//
//    // other redis
//    conn, err := yiigo.Redis("foo").Get()
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    defer yiigo.Redis("foo").Put(conn)
//
//    conn.Do("SET", "test_key", "hello world")
//
// Config
//
//    // yiigo.toml
//    //
//    // [app]
//    // env = "dev"
//    // debug = true
//    // port = 50001
//
//    yiigo.Env("app.env").String("dev")
//    yiigo.Env("app.debug").Bool(true)
//    yiigo.Env("app.port").Int(8000)
//
// HTTP
//
//    c := yiigo.NewHTTPClient(*http.client)
//
//    b, err := c.Get("url...", yiigo.WithRequestTimeout(5*time.Second))
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    fmt.Println(string(b))
//
// Logger
//
//    // default logger
//    yiigo.Logger().Info("hello world")
//
//    // other logger
//    yiigo.Logger("foo").Info("hello world")
//
// For more details, see the documentation for the types and methods.
//
package yiigo
