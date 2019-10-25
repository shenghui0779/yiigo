// Package yiigo makes Golang development easier !
//
// Basic usage
//
// MySQL
//
//    // default db
//    yiigo.DB().Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//    yiigo.Orm().First(&User{}, 1)
//
//    // other db
//    yiigo.DB("foo").Get(&User{}, "SELECT * FROM `user` WHERE `id` = ?", 1)
//    yiigo.Orm("foo").First(&User{}, 1)
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
//    yiigo.Env("app.env").String()
//    yiigo.Env("app.debug").Bool()
//    yiigo.Env("app.port").Int()
//
// Apollo
//
//    type QiniuConfig struct {
//        *yiigo.DefaultApolloConfig
//        BucketName string `toml:"bucket_name"`
//    }
//
//    var qiniu = &QiniuConfig{DefaultApolloConfig: yiigo.NewDefaultConfig("qiniu", "qiniu")}
//
//    if err := yiigo.StartApollo(qiniu); err != nil {
//        log.Fatal(err)
//    }
//
// Zipkin
//
//    reporter := yiigo.NewZipkinHTTPReporter("http://localhost:9411/api/v2/spans")
//
//    // sampler
//    sampler := zipkin.NewModuloSampler(1)
//    // endpoint
//    endpoint, _ := zipkin.NewEndpoint("yiigo-zipkin", "localhost")
//
//    tracer, err := yiigo.NewZipkinTracer(reporter,
//        zipkin.WithLocalEndpoint(endpoint),
//        zipkin.WithSharedSpans(false),
//        zipkin.WithSampler(sampler),
//    )
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    client, err := tracer.HTTPClient(yiigo.WithZipkinClientOptions(zipkinHttp.ClientTrace(true)))
//
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    b, err := client.Get(context.Background(), "url...",
//        yiigo.WithRequestHeader("Content-Type", "application/json"),
//        yiigo.WithRequestTimeout(5*time.Second),
//    )
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
