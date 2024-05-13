package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var controller = map[string]Controller{}

type Controller interface {
	RegisterRouter(r *gin.Engine)
	BasePath() string
}

func Init(e *gin.Engine) {
	for basePath, ctrl := range controller {
		logrus.Debugf("register router %s", basePath)
		ctrl.RegisterRouter(e)
	}
}
