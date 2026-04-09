package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/services/response"
	"time"
)

func init() {
	h := &ClusterHandler{}
	controller[h.BasePath()] = h
}

var _ Controller = &ClusterHandler{}

type ClusterHandler struct{}

func (d *ClusterHandler) BasePath() string {
	return "/v1/cluster"
}

func (d *ClusterHandler) RegisterRouter(r *gin.Engine) {
	group := r.Group(d.BasePath())
	group.GET("/status", d.status)
	group.GET("/diagnostics", d.diagnostics)
	group.GET("/nodes", d.listNodes)
	group.POST("/nodes", d.addNode)
	group.DELETE("/nodes/:nodeId", d.removeNode)
}

func (d *ClusterHandler) status(ctx *gin.Context) {
	response.ResponseData(ctx, cluster.DefaultNode.Status())
}

func (d *ClusterHandler) diagnostics(ctx *gin.Context) {
	diag, err := cluster.DefaultNode.Diagnostics()
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, diag)
}

func (d *ClusterHandler) listNodes(ctx *gin.Context) {
	nodes, err := cluster.DefaultNode.ListMembers()
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, nodes)
}

func (d *ClusterHandler) addNode(ctx *gin.Context) {
	startedAt := time.Now()
	var req cluster.JoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		cluster.DefaultTelemetry.ObserveWrite("cluster_add_member", time.Since(startedAt), "error")
		response.ResponseBadRequest(ctx, err)
		return
	}
	if err := cluster.DefaultNode.AddMember(req); err != nil {
		observeClusterWriteError("cluster_add_member", startedAt, err)
		respondClusterError(ctx, err)
		return
	}
	cluster.DefaultTelemetry.ObserveWrite("cluster_add_member", time.Since(startedAt), "success")
	response.ResponseAccepted(ctx)
}

func (d *ClusterHandler) removeNode(ctx *gin.Context) {
	startedAt := time.Now()
	if err := cluster.DefaultNode.RemoveMember(ctx.Param("nodeId")); err != nil {
		observeClusterWriteError("cluster_remove_member", startedAt, err)
		respondClusterError(ctx, err)
		return
	}
	cluster.DefaultTelemetry.ObserveWrite("cluster_remove_member", time.Since(startedAt), "success")
	response.ResponseAccepted(ctx)
}

func respondClusterError(ctx *gin.Context, err error) {
	if redirect, ok := err.(*cluster.NotLeaderError); ok {
		response.ResponseTemporaryRedirect(ctx, redirect.LeaderAddress)
		return
	}
	response.RespondFailure(ctx, err)
}

func observeClusterWriteError(operation string, startedAt time.Time, err error) {
	outcome := "error"
	if _, ok := err.(*cluster.NotLeaderError); ok {
		outcome = "redirect"
	}
	cluster.DefaultTelemetry.ObserveWrite(operation, time.Since(startedAt), outcome)
}
