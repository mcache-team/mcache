package normal

import "github.com/gin-gonic/gin"

func init() {
	h := &DataHandler{}
	controller[h.BasePath()] = h
}

var _ Controller = &DataHandler{}

type DataHandler struct {
}

func (d *DataHandler) BasePath() string {
	return "/v1/data"
}

func (d *DataHandler) RegisterRouter(e *gin.Engine) {}
