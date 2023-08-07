package router

import (
	e "errors"
	"gfa/core/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewHttpRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.NoRoute(func(c *gin.Context) {
		resp := models.Gin{Ctx: c}
		resp.Fail(http.StatusNotFound, e.New("not found route"))
	})
	router.NoMethod(func(c *gin.Context) {
		resp := models.Gin{Ctx: c}
		resp.Fail(http.StatusNotFound, e.New("not found method"))
	})
	router.GET("/ping", func(c *gin.Context) {
		resp := models.Gin{Ctx: c}
		resp.Success("pong")
	})
	return router
}
