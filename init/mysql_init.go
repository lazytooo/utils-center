package init

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func (config *config) MysqlInit() *sqlx.DB {
	logs := logrus.WithField("envConfig", config)
	maxIdleConn, _ := strconv.Atoi(config.DBMaxIdle)
	maxOpenConn, _ := strconv.Atoi(config.DBMaxOpen)
	if maxIdleConn <= 0 || maxOpenConn <= 0 {
		logs.Errorln("[MySQL] config set err")
		panic("[MySQL] config set err")
	}
	dsnConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_general_ci&loc=Local",
		config.DBUser,
		config.DBPass,
		config.DBAddr,
		config.DBPort,
		config.DBName,
	)

	var err error
	db, err := sqlx.Open("mysql", dsnConfig)
	if err != nil {
		logs.WithError(err).WithField("config", dsnConfig).Errorln("[MySQL] connect err")
		panic("[MySQL] connect err")
	}

	err = db.Ping()
	if err != nil {
		logs.WithError(err).WithField("config", dsnConfig).Errorln("[MySQL] ping pong err")
		panic("[MySQL] ping pong err")
	}

	// 最大连接数限制
	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(time.Minute * 10)

	logrus.Infoln("[MySQL] connect to db success")
	return db
}
