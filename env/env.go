package env

import (
	"github.com/gin-gonic/gin"
	"github.com/go-crt/golib/utils"
	"os"
	"path/filepath"
)

const DefaultRootPath = "."

const (
	// ClusterType 容器中的环境变量
	ClusterType = "CLUSTER_TYPE"
	DockAppName = "APP_NAME"
)

var (
	LocalIP string
	AppName string
	RunMode string

	rootPath        string
	dockerPlateForm bool
)

func init() {
	LocalIP = utils.GetLocalIp()
	dockerPlateForm = false
	RunMode = gin.DebugMode
	if r := os.Getenv(ClusterType); r != "" {
		println("CLUSTER_TYPE=", r)
		dockerPlateForm = true
		RunMode = gin.ReleaseMode
		// 容器里，appName在编排的时候决定
		if n := os.Getenv(DockAppName); n != "" {
			AppName = n
			println("docker env, APP_NAME=", n)
		} else {
			println("docker env, lack APP_NAME!!!")
		}
	}
}

// IsDockerPlatform 判断项目运行平台：容器 vs 开发环境
func IsDockerPlatform() bool {
	return dockerPlateForm
}

// SetAppName 开发环境可手动指定SetAppName
func SetAppName(appName string) {
	if !dockerPlateForm {
		AppName = appName
	}
}

func GetAppName() string {
	return AppName
}

// SetRootPath 设置应用的根目录
func SetRootPath(r string) {
	if !dockerPlateForm {
		rootPath = r
	}
}

// GetRootPath 返回应用的根目录
func GetRootPath() string {
	if rootPath != "" {
		return rootPath
	} else {
		return DefaultRootPath
	}
}

// GetConfDirPath 返回配置文件目录绝对地址
func GetConfDirPath() string {
	return filepath.Join(GetRootPath(), "conf")
}

// GetLogDirPath LogRootPath 返回log目录的绝对地址
func GetLogDirPath() string {
	return filepath.Join(GetRootPath(), "log")
}
