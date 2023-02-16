package env

import (
	"github.com/go-crt/golib/utils"
	"reflect"
	"strings"
)

func getSecret(prefix string) *utils.Conf {
	fileName := strings.TrimRight(strings.TrimLeft(prefix, "@@"), ".")
	secretConf, err := utils.Load(GetConfDirPath()+"/secret/"+fileName+".secret.yaml", nil)
	if err != nil {
		return nil
	}

	return secretConf
}

// CommonSecretChange 目前只支持替换为string类型的value
func CommonSecretChange(prefix string, src, dst interface{}) {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)
	if srcType.Kind() != reflect.Struct || dstType.Kind() != reflect.Ptr {
		return
	}

	secretConf := getSecret(prefix)
	if secretConf == nil {
		return
	}

	// 给dst赋值
	val := reflect.ValueOf(dst).Elem()
	for i := 0; i < srcType.NumField(); i++ {
		if val.Field(i).Kind() != reflect.String {
			continue
		}

		str := val.Field(i).String()
		// 需要替换的secret key只能是string类型
		if strings.HasPrefix(str, prefix) {
			suffix := str[len(prefix):]
			n := secretConf.GetString(suffix)
			if val.Field(i).CanSet() {
				// todo: 原值必为string类型，导致只能替换为string类型
				val.Field(i).SetString(n)
			}
		}
	}
}
