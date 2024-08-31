package handler

import (
	"net/http"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) hwIu9MainPage(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))

	if err != nil {
		// TODO return error
		return
	}
	c.HTML(http.StatusOK, "hw_ui9_mainpage.tmpl", gin.H{
		"Name3": profile.Name3,
	})
}
