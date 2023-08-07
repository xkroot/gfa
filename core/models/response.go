package models

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Gin struct {
	Ctx *gin.Context
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func (g *Gin) Success(data interface{}, args ...string) {
	var msg = "success"
	if len(args) > 0 {
		msg = args[0]
	}
	g.Ctx.JSON(http.StatusOK, Response{
		Code:    200,
		Message: msg,
		Data:    data,
	})
	return
}

func (g *Gin) Fail(httpCode int, err error) {
	g.Ctx.JSON(httpCode, Response{
		Code:    httpCode,
		Message: err.Error(),
	})
	return
}
