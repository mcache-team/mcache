package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/services/response"
	"github.com/mcache-team/mcache/pkg/storage"
)

func init() {
	h := &PrefixHandler{}
	controller[h.BasePath()] = h
}

var _ Controller = &PrefixHandler{}

type PrefixHandler struct {
}

func (d *PrefixHandler) BasePath() string {
	return "/v1/prefix"
}

func (d *PrefixHandler) RegisterRouter(r *gin.Engine) {
	group := r.Group(d.BasePath())
	group.GET("", d.list)
	group.GET("/count", d.count)
}

func (d *PrefixHandler) list(ctx *gin.Context) {
	data, err := storage.StorageClient.ListPrefix("")
	if err != nil {
		response.ResponseNotFound(ctx)
		return
	}
	response.ResponseData(ctx, data)
}

func (d *PrefixHandler) count(ctx *gin.Context) {
	count := storage.StorageClient.CountPrefix("")
	response.ResponseData(ctx, gin.H{"count": count})
}
