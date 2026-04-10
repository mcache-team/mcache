package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/services/response"
	"github.com/mcache-team/mcache/pkg/state"
	"github.com/mcache-team/mcache/pkg/storage"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
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
	group.POST("/:prefix", d.update)
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
	if err := submitWrite(ctx, "data_create", state.Command{
		Type:   state.CommandInsert,
		Prefix: req.Prefix,
		Item: &item.Item{
			Prefix:     req.Prefix,
			Data:       req.Data,
			Timeout:    req.Timeout,
			ExpireTime: req.ExpireTime,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, http.StatusCreated); err != nil {
		return
	}
}

func (d *DataHandler) update(ctx *gin.Context) {
	prefix := ctx.Param("prefix")
	var req item.Item
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ResponseBadRequest(ctx, err)
		return
	}
	if err := submitWrite(ctx, "data_update", state.Command{
		Type:   state.CommandUpdate,
		Prefix: prefix,
		Item: &item.Item{
			Prefix:     prefix,
			Data:       req.Data,
			Timeout:    req.Timeout,
			ExpireTime: req.ExpireTime,
			UpdatedAt:  time.Now(),
		},
	}, http.StatusOK); err != nil {
		return
	}
}

func (d *DataHandler) delete(ctx *gin.Context) {
	prefix := ctx.Param("prefix")
	if err := submitWrite(ctx, "data_delete", state.Command{
		Type:   state.CommandDelete,
		Prefix: prefix,
	}, http.StatusOK); err != nil {
		return
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

func submitWrite(ctx *gin.Context, operation string, cmd state.Command, successStatus int) error {
	startedAt := time.Now()
	_, err := cluster.DefaultNode.Submit(cmd)
	if err == nil {
		cluster.DefaultTelemetry.ObserveWrite(operation, time.Since(startedAt), "success")
		switch successStatus {
		case http.StatusCreated:
			response.ResponseCreated(ctx)
		default:
			response.ResponseSuccess(ctx)
		}
		return nil
	}

	if redirect, ok := err.(*cluster.NotLeaderError); ok {
		cluster.DefaultTelemetry.ObserveWrite(operation, time.Since(startedAt), "redirect")
		response.ResponseTemporaryRedirect(ctx, redirect.LeaderAddress)
		return err
	}
	cluster.DefaultTelemetry.ObserveWrite(operation, time.Since(startedAt), "error")
	if err == item.PrefixExisted {
		response.ResponseAlreadyExists(ctx)
		return err
	}
	response.RespondFailure(ctx, err)
	return err
}
