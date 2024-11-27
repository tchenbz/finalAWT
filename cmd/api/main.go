package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/mailer"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
        port     int
        username string
        password string
        sender   string
    } 
}

type applicationDependencies struct {
	config          serverConfig
	logger          *slog.Logger
	bookModel       data.BookModel
	readingListModel data.ReadingListModel
	reviewModel     data.ReviewModel
	userModel 		data.UserModel
	mailer 			mailer.Mailer
	wg  			sync.WaitGroup 
	tokenModel 		data.TokenModel
	permissionModel data.PermissionModel
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&settings.db.dsn, "db-dsn", os.Getenv("TEST3_DB_DSN"), "PostgreSQL DSN")
	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()
	flag.StringVar(&settings.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&settings.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&settings.smtp.username, "smtp-username", "d120afa780d5b9", "SMTP username")
	flag.StringVar(&settings.smtp.password, "smtp-password", "d72ca97563008b", "SMTP password")
	flag.StringVar(&settings.smtp.sender, "smtp-sender", "Comments Community <no-reply@commentscommunity.tamikachen.net>", "SMTP sender")


	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	appInstance := &applicationDependencies{
		config:          settings,
		logger:          logger,
		bookModel:       data.BookModel{DB: db},
		readingListModel: data.ReadingListModel{DB: db},
		reviewModel:     data.ReviewModel{DB: db},
		userModel: data.UserModel {DB: db},
        mailer: mailer.New(settings.smtp.host, settings.smtp.port,
		settings.smtp.username, settings.smtp.password, settings.smtp.sender),
		tokenModel: data.TokenModel{DB: db},
		permissionModel: data. PermissionModel{DB: db},
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
