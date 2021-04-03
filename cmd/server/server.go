package server

import (
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
	"github.com/kfsoftware/statuspage/pkg/jobs"
	"github.com/pkg/errors"
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

type serverCmd struct {
	config string
	addr   string
}

func (s *serverCmd) validate() error {
	return nil
}
func (s *serverCmd) run() error {
	var err error
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
			return errors.Errorf("Driver %s not supported", driverName)
		}
		dbClient, err = newDbStorage(
			drName,
			dataSource,
		)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("No valid provider: %s", provider)
	}

	r := gin.Default()

	schedRegistry := jobs.NewSchedulerRegistry(
		time.Local,
	)
	log.Infof("Checking all items")
	db.CheckAll(dbClient)
	var checks []db.Check
	result := dbClient.Find(&checks)
	if result.Error != nil {
		return result.Error
	}
	for _, check := range checks {
		chk := check
		duration, err := time.ParseDuration(check.Frecuency)
		if err != nil {
			return err
		}
		err = schedRegistry.Register(chk.ID, duration, func() {
			chk.Check(dbClient)
		})
		if err != nil {
			return err
		}
	}
	es := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			Db:       dbClient,
			Registry: schedRegistry,
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
	listenAddr := s.addr
	if listenAddr == "" {
		listenAddr = viper.GetString("address")
	}
	if listenAddr == "" {
		listenAddr = "0.0.0.0:80"
	}
	err = r.Run(listenAddr)
	if err != nil {
		return err
	}
	return nil
}

func NewServerCmd() *cobra.Command {
	c := &serverCmd{}
	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.SetConfigFile(c.config)
			err := viper.ReadInConfig()
			if err != nil {
				return err
			}
			if err := c.validate(); err != nil {
				return err
			}
			return c.run()
		},
	}
	persistentFlags := cmd.PersistentFlags()
	persistentFlags.StringVarP(&c.config, "config", "", "statuspage", "Configuration file")
	persistentFlags.StringVarP(&c.addr, "address", "", "", "Listen address")

	cmd.MarkPersistentFlagRequired("config")
	return cmd
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
		return nil, err
	}
	err = dbClient.AutoMigrate(&db.CheckExecution{})
	if err != nil {
		return nil, err
	}

	return dbClient, nil
}

type Provider string

const (
	Database Provider = "sql"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
