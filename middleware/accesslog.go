package middleware

import (
	"bytes"
	"fmt"
	"github.com/go-crt/golib/env"
	"github.com/go-crt/golib/utils"
	"github.com/go-crt/golib/xlog"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	printRequestLen  = 10240
	printResponseLen = 10240
)

var (
	// 暂不需要，后续考虑看是否需要支持用户配置
	mcpackReqUris []string
	ignoreReqUris []string
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	s = strings.Replace(s, "\n", "", -1)
	if w.body != nil {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		//idx := len(b)
		// gin render json 后后面会多一个换行符
		//if b[idx-1] == '\n' {
		//	b = b[:idx-1]
		//}
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// access日志打印
func AccessLog() gin.HandlerFunc {
	// 当前模块名
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("%s\n", err)
			}
		}()

		// 开始时间
		start := time.Now()
		// 请求url
		path := c.Request.URL.Path
		// 请求报文
		var requestBody []byte
		if c.Request.Body != nil {
			var err error
			requestBody, err = c.GetRawData()
			if err != nil {
				xlog.Warnf(c, "get http request body error: %s", err.Error())
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		blw := new(bodyLogWriter)
		if printResponseLen <= 0 {
			blw = &bodyLogWriter{body: nil, ResponseWriter: c.Writer}
		} else {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		}
		c.Writer = blw

		c.Set("handler", c.HandlerName())
		logID := xlog.GetLogID(c)
		requestID := xlog.GetRequestID(c)
		// 处理请求
		c.Next()

		response := ""
		if blw.body != nil {
			if len(blw.body.String()) <= printResponseLen {
				response = blw.body.String()
			} else {
				response = blw.body.String()[:printResponseLen]
			}
		}

		bodyStr := ""
		flag := false
		// macpack的请求，以二进制输出日志
		for _, val := range mcpackReqUris {
			if strings.Contains(path, val) {
				bodyStr = fmt.Sprintf("%v", requestBody)
				flag = true
				break
			}
		}
		if !flag {
			// 不打印RequestBody的请求
			for _, val := range ignoreReqUris {
				if strings.Contains(path, val) {
					bodyStr = ""
					flag = true
					break
				}
			}
		}
		if !flag {
			bodyStr = string(requestBody)
		}

		if c.Request.URL.RawQuery != "" {
			bodyStr += "&" + c.Request.URL.RawQuery
		}

		if len(bodyStr) > printRequestLen {
			bodyStr = bodyStr[:printRequestLen]
		}

		// 结束时间
		end := time.Now()

		// 用户自定义notice
		var customerFields []xlog.Field
		for k, v := range xlog.GetCustomerKeyValue(c) {
			customerFields = append(customerFields, xlog.Reflect(k, v))
		}

		// 固定notice
		commonFields := []xlog.Field{
			xlog.String(xlog.TopicType, xlog.LogNameAccess),
			xlog.String("logId", logID),
			xlog.String("requestId", requestID),
			xlog.String("localIp", env.LocalIP),
			xlog.String("module", env.AppName),
			xlog.String("cuid", getReqValueByKey(c, "cuid")),
			xlog.String("device", getReqValueByKey(c, "device")),
			xlog.String("channel", getReqValueByKey(c, "channel")),
			xlog.String("os", getReqValueByKey(c, "os")),
			xlog.String("vc", getReqValueByKey(c, "vc")),
			xlog.String("vcname", getReqValueByKey(c, "vcname")),
			xlog.String("userid", getReqValueByKey(c, "userid")),
			xlog.String("uri", c.Request.RequestURI),
			xlog.String("host", c.Request.Host),
			xlog.String("method", c.Request.Method),
			xlog.String("httpProto", c.Request.Proto),
			xlog.String("handle", c.HandlerName()),
			xlog.String("userAgent", c.Request.UserAgent()),
			xlog.String("refer", c.Request.Referer()),
			xlog.String("clientIp", utils.GetClientIp(c)),
			xlog.String("cookie", getCookie(c)),
			xlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
			xlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
			xlog.Float64("cost", utils.GetRequestCost(start, end)),
			xlog.String("requestParam", bodyStr),
			xlog.Int("responseStatus", c.Writer.Status()),
			xlog.String("response", response),
		}

		commonFields = append(commonFields, customerFields...)
		xlog.InfoLogger(c, "notice", commonFields...)
	}
}

// 从request body中解析特定字段作为notice key打印
func getReqValueByKey(ctx *gin.Context, k string) string {
	if vs, exist := ctx.Request.Form[k]; exist && len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func getCookie(ctx *gin.Context) string {
	cStr := ""
	for _, c := range ctx.Request.Cookies() {
		cStr += fmt.Sprintf("%s=%s&", c.Name, c.Value)
	}
	return strings.TrimRight(cStr, "&")
}

// access 添加kv打印
func AddNotice(k string, v interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		xlog.AddNotice(c, k, v)
		c.Next()
	}
}

//func LoggerBeforeRun(ctx *gin.Context) {
//	customCtx := ctx.CustomContext
//	fields := []xlog.Field{
//		xlog.String("handle", customCtx.HandlerName()),
//		xlog.String("type", customCtx.Type),
//	}
//
//	xlog.InfoLogger(ctx, "start", fields...)
//}

//func LoggerAfterRun(ctx *gin.Context) {
//	customCtx := ctx.CustomContext
//	cost := utils.GetRequestCost(customCtx.StartTime, customCtx.EndTime)
//	var err error
//	if customCtx.Error != nil {
//		err = errors.Cause(customCtx.Error)
//		base.StackLogger(ctx, customCtx.Error)
//	}
//
//	// 用户自定义notice
//	notices := xlog.GetCustomerKeyValue(ctx)
//
//	var fields []xlog.Field
//	for k, v := range notices {
//		fields = append(fields, xlog.Reflect(k, v))
//	}
//
//	fields = append(fields,
//		xlog.String("handle", customCtx.HandlerName()),
//		xlog.String("type", customCtx.Type),
//		xlog.Float64("cost", cost),
//		xlog.String("error", fmt.Sprintf("%+v", err)),
//	)
//
//	xlog.InfoLogger(ctx, "end", fields...)
//}
