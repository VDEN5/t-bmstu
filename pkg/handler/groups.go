package handler

import (
	"net/http"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) groups(c *gin.Context) {
	role := c.GetString("role")

	switch role {
	case "student":
		{
			groups, err := database.GetUserGroups(c.GetString("username"))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}

			c.HTML(http.StatusOK, "groups.tmpl", gin.H{
				"Groups": groups,
			})
		}
	case "teacher":
		{
			c.JSON(http.StatusOK, gin.H{"msg": "Hello, Teacher"})
		}
	case "admin":
		{
			c.JSON(http.StatusOK, gin.H{"msg": "Hello, admin"})
		}
	default:
		{
			c.JSON(http.StatusBadRequest, gin.H{"error": "hacking attempt"})
		}
	}
}
