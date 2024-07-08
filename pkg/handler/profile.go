package handler

import (
	"net/http"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) profileMainPage(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))

	if err != nil {
		// TODO return error
		return
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"NickName": profile.Username,
		"Surname":  profile.LastName,
		"Name":     profile.FirstName,
		"Name3": profile.Name3,
		"Email":    profile.Email,
		//"Avatar":   profile.Avatar,
	})
}
