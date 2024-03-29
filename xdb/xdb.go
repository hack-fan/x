package xdb

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var log *zap.SugaredLogger

// Config 数据库配置，可以被主配置直接引用
type Config struct {
	Host     string `default:"mysql"`
	Port     string `default:"3306"`
	User     string `default:"root"`
	Password string `default:"root"`
	Name     string
	Lifetime int `default:"3000"`
}

func SetLogger(logger *zap.SugaredLogger) {
	log = logger
}

// New 用配置生成一个 gorm mysql 数据库对象,若目标数据库未启动会一直等待
func New(config Config) *gorm.DB {
	var db *gorm.DB
	var err error

	if log == nil {
		zLogger, _ := zap.NewDevelopment()
		log = zLogger.Sugar()
	}

	if config.Name == "" {
		panic("missing db name config")
	}

	var dsn = config.User + ":" + config.Password +
		"@tcp(" + config.Host + ":" + config.Port + ")/" + config.Name +
		"?charset=utf8mb4&parseTime=True&loc=Local&timeout=90s"
	for {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.New(
				zap.NewStdLog(log.Desugar()), // io writer
				logger.Config{
					SlowThreshold:             time.Second * 3, // Slow SQL threshold
					LogLevel:                  logger.Warn,     // Log level
					IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
					Colorful:                  false,           // Disable color
				},
			),
		})
		if err != nil {
			log.Warnw("waiting for connect to db", "origin", err.Error())
			time.Sleep(time.Second * 2)
			continue
		}
		conn, err := db.DB()
		if err != nil {
			panic("can not get db conn from gorm client")
		}
		conn.SetConnMaxLifetime(time.Duration(config.Lifetime) * time.Second)
		log.Info("Mysql connect successful.")
		break
	}

	return db
}
