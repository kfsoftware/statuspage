package main

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kfsoftware/statuspage/pkg/db"
	"github.com/kfsoftware/statuspage/pkg/graphql/generated"
	"github.com/kfsoftware/statuspage/pkg/graphql/resolvers"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"time"
)

var rootCmd = &cobra.Command{
	Use:    "hlf-sync",
	Short:  "HLF sync",
	Long:   `HLF sync is a tool to store all the transaction data of Hyperledger Fabric into a database`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		//cmd.AddCommand(syncCmd)
	},
}

type DriverName string

const (
	PostgresqlDriver = "postgres"
	MySQLDriver      = "mysql"
)

func newDbStorage(driverName DriverName, dataSourceName string) (*gorm.DB, error) {
	var dbClient *gorm.DB
	var err error
	newLogger := logger.New(
		log.New(),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      false,
		},
	)
	gormConfig := &gorm.Config{
		Logger: newLogger,
	}
	switch driverName {
	case PostgresqlDriver:
		dbClient, err = gorm.Open(
			postgres.New(
				postgres.Config{
					DSN:                  dataSourceName,
					PreferSimpleProtocol: true,
				},
			),
			gormConfig,
		)
		if err != nil {
			return nil, err
		}
	case MySQLDriver:
		dbClient, err = gorm.Open(mysql.Open(dataSourceName), gormConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("Driver %s not supported", string(driverName))
	}
	err = dbClient.AutoMigrate(&db.Check{})
	if err != nil {
		panic(err)
	}
	err = dbClient.AutoMigrate(&db.CheckExecution{})
	if err != nil {
		panic(err)
	}

	return dbClient, nil
}

type Provider string

const (
	Database Provider = "sql"
)

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	viper.SetConfigName("statuspage")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("statuspage")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	provider := viper.GetString("database.type")
	var dbClient *gorm.DB
	switch provider {
	case string(Database):
		driverName := viper.GetString("database.driver")
		dataSource := viper.GetString("database.dataSource")
		var drName DriverName
		switch driverName {
		case PostgresqlDriver:
			drName = PostgresqlDriver
		case MySQLDriver:
			drName = MySQLDriver
		default:
			panic(errors.Errorf("Driver %s not supported", driverName))
		}
		dbClient, err = newDbStorage(
			drName,
			dataSource,
		)
		if err != nil {
			panic(err)
		}
	default:
		panic(errors.Errorf("No valid provider: %s", provider))
	}

	r := gin.Default()

	c := cron.New(cron.WithSeconds())
	go func() {
		db.CheckAll(dbClient)
	}()
	spec := viper.GetString("cron")
	if spec == "" {
		spec = "@every 1m"
		log.Warnf("`cron` property not set, defaulting to %s", spec)
	}
	_, err = c.AddFunc(spec, func() {
		db.CheckAll(dbClient)
	})
	if err != nil {
		panic(err)
	}
	c.Start()
	es := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			Db: dbClient,
		},
	})
	h := handler.New(es)

	h.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader:              wsupgrader,
	})
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})
	h.AddTransport(transport.MultipartForm{})

	h.SetQueryCache(lru.New(1000))
	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	h.Use(apollotracing.Tracer{})

	r.Any("/graphql",
		func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Identity")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
			h.ServeHTTP(c.Writer, c.Request)
		},
	)
	playgroundHandler := playground.Handler("GraphQL", "/graphql")
	r.GET("/playground", func(c *gin.Context) {
		playgroundHandler.ServeHTTP(c.Writer, c.Request)
	})
	listenAddr := viper.GetString("address")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:80"
	}
	err = r.Run(listenAddr)
	if err != nil {
		panic(err)
	}
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
