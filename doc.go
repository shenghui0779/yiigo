// Package yiigo makes Golang development easier!
//
// Basic usage
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
//    yiigo.Env("app.env").String()
//    yiigo.Env("app.debug").Bool()
//    yiigo.Env("app.port").Int()
//
// MySQL
//
//    // default db
//    yiigo.DB().Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
//    // other db
//    yiigo.DB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//
// ORM(ent)
//
//    import "<your_project>/ent"
//
//    // default driver
//    client := ent.NewClient(ent.Driver(yiigo.EntDriver()))
//
//    // other driver
//    client := ent.NewClient(ent.Driver(yiigo.EntDriver("other")))
//
// MongoDB
//
//    // default mongodb
//    yiigo.Mongo().Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})
//
//    // other mongodb
//    yiigo.Mongo("foo").Database("test").Collection("numbers").InsertOne(context.Background(), bson.M{"name": "pi", "value": 3.14159})
//
// Redis
//
//    // default redis
//    conn, err := yiigo.Redis().Get(context.Background())
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
//    conn, err := yiigo.Redis("foo").Get(context.Background())
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    defer yiigo.Redis("foo").Put(conn)
//
//    conn.Do("SET", "test_key", "hello world")
//
// HTTP
//
//    // default client
//    yiigo.HTTPGet(context.Background(), "URL", yiigo.WithHTTPTimeout(5*time.Second))
//
//    // new client
//    client := yiigo.NewHTTPClient(*http.Client)
//    client.Get(context.Background(), "URL", yiigo.WithHTTPTimeout(5*time.Second))
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
