package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/services/response"
	"github.com/mcache-team/mcache/pkg/storage"
	"github.com/sirupsen/logrus"
)

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

func (d *DataHandler) RegisterRouter(e *gin.Engine) {
	group := e.Group(d.BasePath())
	group.GET("/listByPrefix", d.listByPrefix)
	group.GET("/:prefix", d.get)
	group.DELETE("/:prefix", d.delete)
	group.PUT("", d.create)
}

func (d *DataHandler) get(ctx *gin.Context) {
	prefix := ctx.Param("prefix")
	data, err := storage.StorageClient.GetOne(prefix)
	if err != nil {
		response.ResponseNotFound(ctx)
		return
	}
	response.ResponseData(ctx, data)
}

func (d *DataHandler) create(ctx *gin.Context) {
	var req item.Item
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Info(req)
		response.ResponseBadRequest(ctx, err)
		return
	}
	if err := handlers.PrefixHandler.InsertNode(req.Prefix, req.Data); err != nil {
		response.ResponseInternalServerError(ctx, err)
		return
	}
	response.ResponseCreated(ctx)
}

func (d *DataHandler) delete(ctx *gin.Context) {
	prefix := ctx.Param("prefix")
	if _, err := storage.StorageClient.Delete(prefix); err != nil {
		response.ResponseInternalServerError(ctx, err)
		return
	} else {
		response.ResponseSuccess(ctx)
	}
}

func (d *DataHandler) listByPrefix(ctx *gin.Context) {
	prefix := ctx.Query("prefix")
	var (
		err  error
		data []*item.Item
	)
	data, err = handlers.PrefixHandler.ListNode(prefix)
	if err != nil {
		response.ResponseInternalServerError(ctx, err)
		return
	}
	response.ResponseData(ctx, data)
}
