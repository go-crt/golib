package base

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/go-crt/golib/env"
	"github.com/go-crt/golib/utils"
	"github.com/go-crt/golib/xlog"
	"reflect"
	"regexp"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const SqlTraceContextKey = "traceContext"

type MysqlConf struct {
	Service         string        `yaml:"service"`
	DataBase        string        `yaml:"database"`
	Addr            string        `yaml:"addr"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	MaxIdleConns    int           `yaml:"maxidleconns"`
	MaxOpenConns    int           `yaml:"maxopenconns"`
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime"`
	ConnTimeOut     time.Duration `yaml:"connTimeOut"`
	WriteTimeOut    time.Duration `yaml:"writeTimeOut"`
	ReadTimeOut     time.Duration `yaml:"readTimeOut"`
	LogMode         bool
}

const prefix = "@@mysql."

func (conf *MysqlConf) checkConf() {
	env.CommonSecretChange(prefix, *conf, conf)

	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 10
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 1000
	}
	if conf.ConnMaxLifeTime == 0 {
		conf.ConnMaxLifeTime = 3600 * time.Second
	}
	if conf.ConnTimeOut == 0 {
		conf.ConnTimeOut = 3 * time.Second
	}
	if conf.WriteTimeOut == 0 {
		conf.WriteTimeOut = 1 * time.Second
	}
	if conf.ReadTimeOut == 0 {
		conf.ReadTimeOut = 1 * time.Second
	}
	// sql 日志为基础的交互日志，默认都打印
	conf.LogMode = true
}

// 日志重定义
var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

type GORMWriter struct {
	LogMode  bool
	Service  string
	Addr     string
	Database string
}

func (gw GORMWriter) Print(values ...interface{}) {
	if gw.LogMode {
		msg, fields := gormLogKeyValueFormatter(values...)
		if len(values) > 1 && len(fields) > 0 {
			end := time.Now()
			fields = append(fields,
				xlog.String("module", env.GetAppName()),
				xlog.String("service", gw.Service),
				xlog.String("addr", gw.Addr),
				xlog.String("db", gw.Database),
				xlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
			)

			if values[0] == "sql" && len(values) > 3 {
				startNs := end.UnixNano() - values[2].(time.Duration).Nanoseconds()
				start := time.Unix(startNs/1e9, startNs*1.0%1e9)
				fields = append(fields, xlog.String("requestStartTime", utils.GetFormatRequestTime(start)))

			}
			xlog.InfoLogger(nil, msg, fields...)
		}
	}
}

func gormLogKeyValueFormatter(values ...interface{}) (msg string, fields []xlog.Field) {
	if len(values) <= 1 {
		return
	}
	if values[0] == "sql" {
		var sql string
		var formattedValues []string

		// sql
		for _, value := range values[4].([]interface{}) {
			indirectValue := reflect.Indirect(reflect.ValueOf(value))
			if indirectValue.IsValid() {
				value = indirectValue.Interface()
				if t, ok := value.(time.Time); ok {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
				} else if b, ok := value.([]byte); ok {
					if str := string(b); isPrintable(str) {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
					} else {
						formattedValues = append(formattedValues, "'<binary>'")
					}
				} else if r, ok := value.(driver.Valuer); ok {
					if value, err := r.Value(); err == nil && value != nil {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					} else {
						formattedValues = append(formattedValues, "NULL")
					}
				} else {
					switch value.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
						formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
					default:
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				}
			} else {
				formattedValues = append(formattedValues, "NULL")
			}
		}

		// differentiate between $n placeholders or else treat like ?
		if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
			sql = values[3].(string)
			for index, value := range formattedValues {
				placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
				sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
			}
		} else {
			formattedValuesLength := len(formattedValues)
			for index, value := range sqlRegexp.Split(values[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}
		}

		var logId string
		var requestId string
		if len(values) >= 7 && values[6] != nil {
			logId, _ = values[6].(context.Context).Value(xlog.ContextKeyLogID).(string)
			requestId, _ = values[6].(context.Context).Value(xlog.ContextKeyRequestID).(string)
		}
		fields = []xlog.Field{
			xlog.String(xlog.TopicType, xlog.LogNameModule),
			xlog.String("logId", logId),
			xlog.String("requestId", requestId),
			xlog.String("sql", sql),
			xlog.Float64("cost", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0),
			xlog.Int64("affectedrow", values[5].(int64)),
			xlog.Int("ralCode", 0),
			xlog.String("prot", "mysql"),
		}

		// todo: 这里打印的日志并不代表真的成功。比如 table doesn't exist 会先打印一条日志，然后输出本sql语句
		msg := "mysql do success"
		return msg, fields
	}

	if values[0] == "log" {
		ctx := values[1]
		fileLineNum := values[2]
		vars := values[3].([]interface{})
		var logId string
		var requestId string
		if ctx != nil {
			logId, _ = ctx.(context.Context).Value(xlog.ContextKeyLogID).(string)
			requestId, _ = ctx.(context.Context).Value(xlog.ContextKeyRequestID).(string)
		}
		var msg interface{}
		msg = values[3]
		if len(vars) == 1 {
			if m, ok := vars[0].(error); ok {
				msg = m.Error()
			}
		}
		fields = []xlog.Field{
			xlog.String(xlog.TopicType, xlog.LogNameModule),
			xlog.Reflect("file", fileLineNum),
			xlog.String("logId", logId),
			xlog.String("requestId", requestId),
			xlog.Int("ralCode", -1),
			xlog.String("prot", "mysql"),
		}
		return msg.(string), fields
	}
	return
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func InitMysqlClient(conf MysqlConf) (client *gorm.DB, err error) {
	conf.checkConf()

	client, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=True&loc=Asia%%2FShanghai",
		conf.User,
		conf.Password,
		conf.Addr,
		conf.DataBase,
		conf.ConnTimeOut,
		conf.ReadTimeOut,
		conf.WriteTimeOut))

	if err != nil {
		return client, err
	}

	client.DB().SetMaxIdleConns(conf.MaxIdleConns)
	client.DB().SetMaxOpenConns(conf.MaxOpenConns)
	client.DB().SetConnMaxLifetime(conf.ConnMaxLifeTime)
	client.LogMode(conf.LogMode)

	ormLogger := GORMWriter{
		LogMode:  conf.LogMode,
		Addr:     conf.Addr,
		Database: conf.DataBase,
	}
	if conf.Service != "" {
		ormLogger.Service = conf.Service
	} else {
		ormLogger.Service = conf.DataBase
	}
	client.SetLogger(ormLogger)

	// register tracer callback
	setCallback(client, "create")
	setCallback(client, "delete")
	setCallback(client, "update")
	setCallback(client, "query")
	setCallback(client, "row_query")

	return client, nil
}

func setCallback(client *gorm.DB, callbackName string) {
	beforeName := fmt.Sprintf("tracer:%v_before", callbackName)
	afterName := fmt.Sprintf("tracer:%v_after", callbackName)
	gormCallbackName := fmt.Sprintf("gorm:%v", callbackName)
	switch callbackName {
	case "create":
		client.Callback().Create().Before(gormCallbackName).Register(beforeName, func(scope *gorm.Scope) {
			tracerBefore(scope, callbackName)
		})
		client.Callback().Create().After(gormCallbackName).Register(afterName, func(scope *gorm.Scope) {
			tracerAfter(scope, callbackName)
		})
	case "query":
		client.Callback().Query().Before(gormCallbackName).Register(beforeName, func(scope *gorm.Scope) {
			tracerBefore(scope, callbackName)
		})
		client.Callback().Query().After(gormCallbackName).Register(afterName, func(scope *gorm.Scope) {
			tracerAfter(scope, callbackName)
		})
	case "update":
		client.Callback().Update().Before(gormCallbackName).Register(beforeName, func(scope *gorm.Scope) {
			tracerBefore(scope, callbackName)
		})
		client.Callback().Update().After(gormCallbackName).Register(afterName, func(scope *gorm.Scope) {
			tracerAfter(scope, callbackName)
		})
	case "delete":
		client.Callback().Delete().Before(gormCallbackName).Register(beforeName, func(scope *gorm.Scope) {
			tracerBefore(scope, callbackName)
		})
		client.Callback().Delete().After(gormCallbackName).Register(afterName, func(scope *gorm.Scope) {
			tracerAfter(scope, callbackName)
		})
	case "row_query":
		client.Callback().RowQuery().Before(gormCallbackName).Register(beforeName, func(scope *gorm.Scope) {
			tracerBefore(scope, callbackName)
		})
		client.Callback().RowQuery().After(gormCallbackName).Register(afterName, func(scope *gorm.Scope) {
			tracerAfter(scope, callbackName)
		})
	}
}

func tracerBefore(scope *gorm.Scope, callbackName string) {
}

func tracerAfter(scope *gorm.Scope, callbackName string) {
}
