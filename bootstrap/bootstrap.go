package bootstrap

import (
	"flag"
	"fmt"
	"gin-api/global"
	"gin-api/jobs"
	"gin-api/libraries/config"
	"gin-api/libraries/goredis"
	"gin-api/libraries/jaeger"
	"gin-api/libraries/logging"
	"gin-api/libraries/mysql"
	"gin-api/libraries/redis"
	"gin-api/resource"
	"gin-api/routers"
	"log"
	"net/http"
	"syscall"
	"time"
)

var (
	conf = flag.String("conf", "conf_dev", "config path")
	job  = flag.String("job", "", "is job")
)

func Bootstrap() {
	flag.Parse()

	initConfig()
	initApp()
	initLogger()
	initMysql("test_mysql")
	initRedis("default_redis")
	initJaeger()

	if *job == "" {
		log.Println("start by server")
		initHTTP()
	} else {
		jobs.Handle(*job)
	}
}

func initConfig() {
	log.Println("The conf path is :" + *conf)
	var err error
	resource.Config = config.InitConfig(*conf, "toml")
	if err != nil {
		panic(err)
	}
}

func initApp() {
	if err := resource.Config.ReadConfig("app", "toml", &global.Global); err != nil {
		panic(err)
	}
}

func initLogger() {
	var (
		err error
		cfg logging.Config
	)

	if err = resource.Config.ReadConfig("log", "toml", &cfg); err != nil {
		panic(err)
	}

	resource.Logger = logging.NewLogger(cfg)
}

func initMysql(db string) {
	var (
		err error
		cfg mysql.Config
	)

	if err = resource.Config.ReadConfig(db, "toml", &cfg); err != nil {
		panic(err)
	}

	resource.TestDB, err = mysql.NewMySQL(cfg)
	if err != nil {
		panic(err)
	}
}

func initRedis(db string) {
	var (
		err   error
		cfg   redis.Config
		goCfg goredis.Config
	)

	if err = resource.Config.ReadConfig(db, "toml", &cfg); err != nil {
		panic(err)
	}

	if err = resource.Config.ReadConfig(db, "toml", &goCfg); err != nil {
		panic(err)
	}

	resource.DefaultRedis, err = redis.GetRedis(cfg)
	if err != nil {
		panic(err)
	}

	rc := goredis.NewClient(&goCfg)
	rc.AddHook(jaeger.NewJaegerHook())
	resource.GoRedis = rc
}

func initJaeger() {
	var (
		err error
		cfg jaeger.Config
	)

	if err = resource.Config.ReadConfig("jaeger", "toml", &cfg); err != nil {
		panic(err)
	}

	_, _, err = jaeger.NewJaegerTracer(cfg)
	if err != nil {
		panic(err)
	}

	return
}

type HTTPConfig struct {
	Port int
}

func initHTTP() {
	router := routers.InitRouter()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", global.Global.AppPort),
		Handler:      router,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	log.Printf("Actual pid is %d", syscall.Getpid())
	log.Printf("Actual port is %d", global.Global.AppPort)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

	// endless.DefaultReadTimeOut = 3 * time.Second
	// endless.DefaultWriteTimeOut = 3 * time.Second
	// serverEnd := endless.NewServer(fmt.Sprintf(":%d", global.Global.AppPort), router)
	// err = serverEnd.ListenAndServe()
	// if err != nil {
	// 	log.Printf("Server err: %v", err)
	// }
}
