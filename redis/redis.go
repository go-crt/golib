package redis

import (
	"fmt"
	"github.com/go-crt/golib/utils"
	"github.com/go-crt/golib/xlog"
	"time"

	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
)

// 日志打印Do args部分支持的最大长度
const logForRedisValue = 50

type RedisConf struct {
	Service      string        `yaml:"service"`
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	MaxIdle      int           `yaml:"maxIdle"`
	MaxActive    int           `yaml:"maxActive"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
	ConnTimeOut  time.Duration `yaml:"connTimeOut"`
	ReadTimeOut  time.Duration `yaml:"readTimeOut"`
	WriteTimeOut time.Duration `yaml:"writeTimeOut"`
}

func (conf *RedisConf) checkConf() {
	if conf.MaxIdle == 0 {
		conf.MaxIdle = 100
	}
	if conf.MaxActive == 0 {
		conf.MaxActive = 100
	}
	if conf.IdleTimeout == 0 {
		conf.IdleTimeout = 10 * time.Millisecond
	}
	if conf.ConnTimeOut == 0 {
		conf.ConnTimeOut = 1 * time.Second
	}
	if conf.ReadTimeOut == 0 {
		conf.ReadTimeOut = 1 * time.Second
	}
	if conf.WriteTimeOut == 0 {
		conf.WriteTimeOut = 1 * time.Second
	}
}

// Redis 日志打印Do args部分支持的最大长度
type Redis struct {
	pool       *redigo.Pool
	Service    string
	RemoteAddr string
}

func InitRedisClient(conf RedisConf) (*Redis, error) {
	conf.checkConf()
	p := &redigo.Pool{
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: conf.IdleTimeout,
		Wait:        true,
		Dial: func() (conn redigo.Conn, e error) {
			con, err := redigo.Dial(
				"tcp",
				conf.Addr,
				redigo.DialPassword(conf.Password),
				redigo.DialConnectTimeout(conf.ConnTimeOut),
				redigo.DialReadTimeout(conf.ReadTimeOut),
				redigo.DialWriteTimeout(conf.WriteTimeOut),
			)
			if err != nil {
				return nil, err
			}
			return con, nil
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	c := &Redis{
		Service:    conf.Service,
		RemoteAddr: conf.Addr,
		pool:       p,
	}
	return c, nil
}

func (r *Redis) Do(ctx *gin.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	start := time.Now()

	conn := r.pool.Get()
	err = conn.Err()
	if err != nil {
		xlog.ErrorLogger(ctx, "get connection error: "+err.Error(), xlog.String("prot", "redis"))
		return reply, err
	}

	reply, err = conn.Do(commandName, args...)
	if err = conn.Close(); err != nil {
		xlog.WarnLogger(ctx, "connection close error: "+err.Error(), xlog.String("prot", "redis"))
	}

	end := time.Now()

	// 执行时间 单位:毫秒
	ralCode := 0
	msg := "redis do success"
	if err != nil {
		ralCode = -1
		msg = fmt.Sprintf("redis do error: %s", err.Error())
		xlog.ErrorLogger(ctx, msg, xlog.String("prot", "redis"))
	}

	fields := []xlog.Field{
		xlog.String(xlog.TopicType, xlog.LogNameModule),
		xlog.String("prot", "redis"),
		xlog.String("remoteAddr", r.RemoteAddr),
		xlog.String("service", r.Service),
		xlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		xlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		xlog.Float64("cost", utils.GetRequestCost(start, end)),
		xlog.String("command", commandName),
		xlog.String("commandVal", utils.JoinArgs(logForRedisValue, args)),
		xlog.Int("ralCode", ralCode),
	}

	xlog.InfoLogger(ctx, msg, fields...)
	return reply, err
}

func (r *Redis) Close() error {
	return r.pool.Close()
}

func (r *Redis) Stats() (inUseCount, idleCount, activeCount int) {
	stats := r.pool.Stats()
	idleCount = stats.IdleCount
	activeCount = stats.ActiveCount
	inUseCount = activeCount - idleCount
	return inUseCount, idleCount, activeCount
}
