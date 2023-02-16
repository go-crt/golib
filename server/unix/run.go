package unix

import (
	"os"

	"github.com/gin-gonic/gin"
)

const (
	socketPath = "/usr/local/var/run/"
	sockName   = "go.sock"
)

func Start(router *gin.Engine) {
	var err error
	go func() {
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			err = os.MkdirAll(socketPath, os.ModePerm)
			if err != nil {
				panic("mkdir " + socketPath + "error: " + err.Error())
			}
		}
		err = router.RunUnix(socketPath + sockName)
		if err != nil {
			panic("runUnix error: " + err.Error())
		}
	}()
}
