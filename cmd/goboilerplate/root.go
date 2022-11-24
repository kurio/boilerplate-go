package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kurio/boilerplate-go/cmd/logger"
	_config "github.com/kurio/boilerplate-go/internal/config"
)

// TODO: update
const app = "goboilerplate"

var (
	rootCMD = &cobra.Command{
		Use:   app,
		Short: "Short description.",
	}

	versionCMD = &cobra.Command{
		Use:   "version",
		Short: "Get binary version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: '%s'\n", gitCommit)
		},
	}

	config    _config.Config
	gitCommit string

	mysqlDB     *sql.DB
	mongoClient *mongo.Client
	redisClient redis.UniversalClient
	httpClient  *http.Client
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCMD.AddCommand(versionCMD)
	rootCMD.PersistentFlags().String("config", "", "Set this flag to use a configuration file.")
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.BindPFlag("config", rootCMD.Flags().Lookup("config")); err != nil {
		logrus.Fatalf("error binding pflag 'config': %+v", err)
	}

	if configFile := viper.GetString("config"); configFile != "" {
		logrus.Infof("using configFile: %s", configFile)
		viper.SetConfigType("json")
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			logrus.Errorf("error reading config file '%s': %+v", viper.ConfigFileUsed(), err)
		}
	}

	// setConfig()
	config = _config.LoadConfig()

	logger.SetupLogs(config.LogLevelStr)
	if config.Debug {
		logrus.Warn("app is running in debug mode")
	} else {
		logrus.Info("app is running in production mode")
	}
}

func initMysqlDB() {
	var err error

	mysqlDB, err = sql.Open("mysql", config.MySQL.DSN)
	if err != nil {
		logrus.Fatalf("Failed to initialize mysql client: %+v", err)
	}
	mysqlDB.SetConnMaxLifetime(config.MySQL.ConnMaxLifetime)
	mysqlDB.SetMaxIdleConns(config.MySQL.MaxIdleConns)
	mysqlDB.SetMaxOpenConns(config.MySQL.MaxOpenConns)
}

func initMongoClient() {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoClient, err = mongo.Connect(
		ctx,
		options.Client().
			ApplyURI(config.Mongo.URI).
			SetConnectTimeout(config.Mongo.ConnectTimeout).
			SetServerSelectionTimeout(config.Mongo.ServerSelectionTimeout),
	)
	if err != nil {
		logrus.Fatalf("Failed to initialize mongodb client: %+v", err)
	}

	err = mongoClient.Ping(context.Background(), nil)
	if err != nil {
		logrus.Fatalf("error pinging mongodb: %+v", err)
	}
}

func initRedisClient() {
	if config.Redis.Cluster {
		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           strings.Split(config.Redis.Address, ","),
			DialTimeout:     config.Redis.DialTimeout,
			ReadTimeout:     config.Redis.ReadTimeout,
			WriteTimeout:    config.Redis.WriteTimeout,
			PoolSize:        config.Redis.PoolSize,
			PoolTimeout:     config.Redis.PoolTimeout,
			ConnMaxIdleTime: config.Redis.ConnMaxIdleTime,
			ConnMaxLifetime: config.Redis.ConnMaxLifetime,
		})
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:            config.Redis.Address,
			DialTimeout:     config.Redis.DialTimeout,
			ReadTimeout:     config.Redis.ReadTimeout,
			WriteTimeout:    config.Redis.WriteTimeout,
			PoolSize:        config.Redis.PoolSize,
			PoolTimeout:     config.Redis.PoolTimeout,
			ConnMaxIdleTime: config.Redis.ConnMaxIdleTime,
			ConnMaxLifetime: config.Redis.ConnMaxLifetime,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.Fatalf("error pinging redis: %+v", err)
	}
}

func initHttpClient() {
	defaultTransport := &http.Transport{
		MaxIdleConns:        config.HTTP.Client.MaxIdleConns,
		MaxIdleConnsPerHost: config.HTTP.Client.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.HTTP.Client.IdleConnTimeout,
	}

	httpClient = new(http.Client)
	httpClient.Timeout = config.HTTP.Client.Timeout
	httpClient.Transport = defaultTransport
}
