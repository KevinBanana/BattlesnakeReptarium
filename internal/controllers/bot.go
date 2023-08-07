package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type BotController struct{}

func (b BotController) GetMove(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}
