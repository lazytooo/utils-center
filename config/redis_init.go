package config

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"time"
)

func (config *config) RedisInit() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPass,
		DB:       0, // use default DB
		PoolSize: 300,
		// 最低维持连接数
		MinIdleConns: 3,
		PoolTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		logrus.WithError(err).WithField("pong", pong).Fatalln("[Redis] connect err")
		panic("[Redis] connect err")
	}

	logrus.Infoln("[Redis] connect to redis success")
	return client
}
