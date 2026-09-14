package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/services/response"
)

var memoryHandler = handlers.Memory

func init() {
	h := &MemoryHandler{}
	controller[h.BasePath()] = h
}

var _ Controller = &MemoryHandler{}

type MemoryHandler struct{}

func (d *MemoryHandler) BasePath() string {
	return "/v1/memory"
}

func (d *MemoryHandler) RegisterRouter(e *gin.Engine) {
	group := e.Group(d.BasePath())
	group.POST("/agents/:agentId/sessions", d.createSession)
	group.GET("/agents/:agentId/sessions", d.listSessions)
	group.DELETE("/agents/:agentId/sessions/:sessionId", d.deleteSession)
	group.POST("/agents/:agentId/sessions/:sessionId/working", d.appendWorking)
	group.GET("/agents/:agentId/sessions/:sessionId/working", d.listWorking)
	group.PUT("/agents/:agentId/sessions/:sessionId/scratchpad/:key", d.putScratchpad)
	group.GET("/agents/:agentId/sessions/:sessionId/scratchpad", d.listScratchpad)
	group.GET("/agents/:agentId/sessions/:sessionId/scratchpad/:key", d.getScratchpad)
	group.POST("/agents/:agentId/facts", d.putFact)
	group.GET("/agents/:agentId/facts", d.listFacts)
	group.GET("/agents/:agentId/facts/:key", d.getFact)
	group.DELETE("/agents/:agentId/facts/:key", d.deleteFact)
}

type createSessionRequest struct {
	SessionID  string `json:"sessionId,omitempty"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func (d *MemoryHandler) createSession(ctx *gin.Context) {
	var req createSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ResponseBadRequest(ctx, err)
		return
	}
	sessionID, err := memoryHandler.CreateSession(ctx.Param("agentId"), req.SessionID, req.TTLSeconds)
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, gin.H{"agentId": ctx.Param("agentId"), "sessionId": sessionID})
}

func (d *MemoryHandler) listSessions(ctx *gin.Context) {
	sessions, err := memoryHandler.ListSessions(ctx.Param("agentId"))
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, sessions)
}

func (d *MemoryHandler) deleteSession(ctx *gin.Context) {
	removed, err := memoryHandler.DeleteSession(ctx.Param("agentId"), ctx.Param("sessionId"))
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, gin.H{"removed": removed})
}

type appendWorkingRequest struct {
	Content    string `json:"content"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func (d *MemoryHandler) appendWorking(ctx *gin.Context) {
	var req appendWorkingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ResponseBadRequest(ctx, err)
		return
	}
	entryID, err := memoryHandler.AppendWorkingMemory(ctx.Param("agentId"), ctx.Param("sessionId"), req.Content, req.TTLSeconds)
	if err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseData(ctx, gin.H{"entryId": entryID})
}

func (d *MemoryHandler) listWorking(ctx *gin.Context) {
	items, err := memoryHandler.ListWorkingMemory(ctx.Param("agentId"), ctx.Param("sessionId"))
	if err != nil {
		response.ResponseData(ctx, []interface{}{})
		return
	}
	response.ResponseData(ctx, items)
}

type scratchpadRequest struct {
	Content    string `json:"content"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func (d *MemoryHandler) putScratchpad(ctx *gin.Context) {
	var req scratchpadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ResponseBadRequest(ctx, err)
		return
	}
	if err := memoryHandler.PutScratchpad(ctx.Param("agentId"), ctx.Param("sessionId"), ctx.Param("key"), req.Content, req.TTLSeconds); err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseSuccess(ctx)
}

func (d *MemoryHandler) getScratchpad(ctx *gin.Context) {
	it, err := memoryHandler.GetScratchpad(ctx.Param("agentId"), ctx.Param("sessionId"), ctx.Param("key"))
	if err != nil {
		response.ResponseNotFound(ctx)
		return
	}
	response.ResponseData(ctx, it)
}

func (d *MemoryHandler) listScratchpad(ctx *gin.Context) {
	items, err := memoryHandler.ListScratchpad(ctx.Param("agentId"), ctx.Param("sessionId"))
	if err != nil {
		response.ResponseData(ctx, []interface{}{})
		return
	}
	response.ResponseData(ctx, items)
}

type putFactRequest struct {
	Key        string `json:"key"`
	Content    string `json:"content"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func (d *MemoryHandler) putFact(ctx *gin.Context) {
	var req putFactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ResponseBadRequest(ctx, err)
		return
	}
	if err := memoryHandler.PutFact(ctx.Param("agentId"), req.Key, req.Content, req.TTLSeconds); err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseSuccess(ctx)
}

func (d *MemoryHandler) listFacts(ctx *gin.Context) {
	items, err := memoryHandler.GetFacts(ctx.Param("agentId"), "")
	if err != nil {
		response.ResponseData(ctx, []interface{}{})
		return
	}
	response.ResponseData(ctx, items)
}

func (d *MemoryHandler) getFact(ctx *gin.Context) {
	items, err := memoryHandler.GetFacts(ctx.Param("agentId"), ctx.Param("key"))
	if err != nil || len(items) == 0 {
		response.ResponseNotFound(ctx)
		return
	}
	response.ResponseData(ctx, items[0])
}

func (d *MemoryHandler) deleteFact(ctx *gin.Context) {
	if err := memoryHandler.DeleteFact(ctx.Param("agentId"), ctx.Param("key")); err != nil {
		response.RespondFailure(ctx, err)
		return
	}
	response.ResponseSuccess(ctx)
}
