package response

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"net/http"
	"strings"
	"time"
)

var (
	successResp = gin.H{"code": 0, "msg": "OK"}
)

type HttpError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	TimeStamp int64  `json:"timestamp"`
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("%s [%d] - %s", time.Now().Format(time.DateTime), e.Code, e.Message)
}

func ResponseData(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

func ResponseCreated(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, successResp)
}

func ResponseSuccess(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, successResp)
}

func ResponseAccepted(ctx *gin.Context) {
	ctx.JSON(http.StatusAccepted, successResp)
}

func ResponseBadRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, NewErrorResp(http.StatusBadGateway, err.Error()))
}

func ResponseNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, NewErrorResp(http.StatusNotFound, "cache not found"))
}

func ResponseAlreadyExists(ctx *gin.Context) {
	ctx.JSON(http.StatusConflict, NewErrorResp(http.StatusConflict, "cache already exists"))
}

func ResponseInternalServerError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusInternalServerError, NewErrorResp(http.StatusInternalServerError, err.Error()))
}

func RespondFailure(ctx *gin.Context, err error) {
	if strings.Contains(err.Error(), item.PrefixNotExisted.Error()) {
		ctx.JSON(http.StatusNotFound, NewErrorResp(http.StatusNotFound, "cache not found"))
		return
	}

	if response, ok := err.(*HttpError); ok {
		ctx.JSON(response.Code, response)
		return
	}

	ResponseInternalServerError(ctx, err)
}

func NewErrorResp(code int, err string) *HttpError {
	return &HttpError{
		Code:      code,
		Message:   err,
		TimeStamp: time.Now().Unix(),
	}
}
