package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrInternalService = errors.New("internal service error")
	ErrBadRequest      = errors.New("bad request")
)

type Handler struct {
	env      string
	validate *validator.Validate
}

type Bundle struct {
	Env string
}

func New(b Bundle) *Handler {
	v := validator.New()

	return &Handler{
		env:      b.Env,
		validate: v,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {

}
