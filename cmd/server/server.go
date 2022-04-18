package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/kfsoftware/statuspage/config"
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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	_ "net/http/pprof"
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
	viper.SetDefault("database.type", "sql")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dataSource", "gorm.db")
	provider := viper.GetString("database.type")
	var dbClient *gorm.DB
	var drName config.DriverName
	switch provider {
	case string(Database):
		driverName := viper.GetString("database.driver")
		dataSource := viper.GetString("database.dataSource")
		switch driverName {
		case config.PostgresqlDriver:
			drName = config.PostgresqlDriver
		case config.MySQLDriver:
			drName = config.MySQLDriver
		case config.SQLiteDriver:
			drName = config.SQLiteDriver
		default:
			drName = config.SQLiteDriver
			log.Warnf("No database driver specified, defaulting to %s", config.SQLiteDriver)
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
			DriverName: drName,
			Db:         dbClient,
			Registry:   schedRegistry,
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

	graphqlHandler := func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Identity")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		h.ServeHTTP(writer, request)
	}
	http.HandleFunc(
		"/graphql",
		graphqlHandler,
	)
	playgroundHandler := playground.Handler("GraphQL", "/graphql")
	http.HandleFunc(
		"/playground",
		playgroundHandler,
	)
	listenAddr := s.addr
	if listenAddr == "" {
		listenAddr = viper.GetString("address")
	}
	if listenAddr == "" {
		listenAddr = "0.0.0.0:80"
	}
	err = http.ListenAndServe(listenAddr, nil)
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
			if c.config != "" {
				viper.SetConfigFile(c.config)
				err := viper.ReadInConfig()
				if err != nil {
					return err
				}
			}
			if err := c.validate(); err != nil {
				return err
			}
			return c.run()
		},
	}
	persistentFlags := cmd.PersistentFlags()
	persistentFlags.StringVarP(&c.config, "config", "", "", "Configuration file")
	persistentFlags.StringVarP(&c.addr, "address", "", "", "Listen address")
	return cmd
}

func newDbStorage(driverName config.DriverName, dataSourceName string) (*gorm.DB, error) {
	var dbClient *gorm.DB
	var err error
	newLogger := logger.New(
		log.StandardLogger(),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      false,
		},
	).LogMode(logger.Info)
	gormConfig := &gorm.Config{
		Logger:               newLogger,
		FullSaveAssociations: true,
	}
	switch driverName {
	case config.PostgresqlDriver:
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
	case config.MySQLDriver:
		dbClient, err = gorm.Open(mysql.Open(dataSourceName), gormConfig)
		if err != nil {
			return nil, err
		}
	case config.SQLiteDriver:
		dbClient, err = gorm.Open(sqlite.Open(dataSourceName), gormConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("Driver %s not supported", string(driverName))
	}
	err = dbClient.AutoMigrate(&db.PageCheck{})
	if err != nil {
		return nil, err
	}
	err = dbClient.AutoMigrate(&db.Check{})
	if err != nil {
		return nil, err
	}
	err = dbClient.AutoMigrate(&db.CheckExecution{})
	if err != nil {
		return nil, err
	}
	err = dbClient.AutoMigrate(&db.StatusPage{})
	if err != nil {
		return nil, err
	}
	err = dbClient.AutoMigrate(&db.Namespace{})
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
