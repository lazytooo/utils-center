package init

import "os"

type config struct {
	// 数据库配置
	DBAddr    string // 数据库地址
	DBPort    string // 数据库端口
	DBName    string // 数据库名称
	DBUser    string // 数据库登录用户名
	DBPass    string // 数据库登录密码
	DBMaxIdle string // 数据库最大连接数
	DBMaxOpen string // 数据库最大开启连接数

	// Redis配置
	RedisAddr string // Redis地址
	RedisPass string // Redis密码

	// Websocket配置
	WebsocketAddr string // websocket地址

	// Kafka配置
}

func NewConfig() *config {
	return &config{
		DBAddr:        os.Getenv("DB_ADDR"),
		DBPort:        os.Getenv("DB_PORT"),
		DBName:        os.Getenv("DB_NAME"),
		DBUser:        os.Getenv("DB_USER"),
		DBPass:        os.Getenv("DB_PASS"),
		DBMaxIdle:     os.Getenv("DB_MAX_IDLE"),
		DBMaxOpen:     os.Getenv("DB_MAX_OPEN"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPass:     os.Getenv("REDIS_PASS"),
		WebsocketAddr: os.Getenv("WS_ADDR"),
	}
}
