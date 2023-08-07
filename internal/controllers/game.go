package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GameController struct{}

func (g GameController) StartGame(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) EndGame(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) CalculateMove(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"time": time.Now()})
}
