package redis

import (
	"errors"
	"github.com/go-crt/golib/utils"
	"github.com/go-crt/golib/xlog"
	"time"

	"github.com/gin-gonic/gin"
)

type Pipeliner interface {
	Exec(ctx *gin.Context) ([]interface{}, error)
	Put(ctx *gin.Context, cmd string, args ...interface{}) error
}

type commands struct {
	cmd   string
	args  []interface{}
	reply interface{}
	err   error
}

type Pipeline struct {
	cmds  []commands
	err   error
	redis *Redis
}

func (r *Redis) Pipeline() Pipeliner {
	return &Pipeline{
		redis: r,
	}
}

func (p *Pipeline) Put(ctx *gin.Context, cmd string, args ...interface{}) error {
	if len(args) < 1 {
		return errors.New("no key found in args")
	}
	c := commands{
		cmd:  cmd,
		args: args,
	}
	p.cmds = append(p.cmds, c)
	return nil
}

func (p *Pipeline) Exec(ctx *gin.Context) (res []interface{}, err error) {
	start := time.Now()

	conn := p.redis.pool.Get()
	defer conn.Close()

	for i := range p.cmds {
		err = conn.Send(p.cmds[i].cmd, p.cmds[i].args...)
	}

	err = conn.Flush()

	var msg string
	var ralCode int
	if err == nil {
		ralCode = 0
		for i := range p.cmds {
			var reply interface{}
			reply, err = conn.Receive()
			res = append(res, reply)
			p.cmds[i].reply, p.cmds[i].err = reply, err
		}

		msg = "pipeline exec succ"
	} else {
		ralCode = -1
		p.err = err
		msg = "pipeline exec error: " + err.Error()
	}

	end := time.Now()
	fields := []xlog.Field{
		xlog.String(xlog.TopicType, xlog.LogNameModule),
		xlog.String("prot", "redis"),
		xlog.String("remoteAddr", p.redis.RemoteAddr),
		xlog.String("service", p.redis.Service),
		xlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		xlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		xlog.Float64("cost", utils.GetRequestCost(start, end)),
		xlog.Int("ralCode", ralCode),
	}

	xlog.InfoLogger(ctx, msg, fields...)

	return res, err
}
