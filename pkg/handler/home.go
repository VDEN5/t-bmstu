package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) add(c *gin.Context) {
	members := []json.RawMessage{
		json.RawMessage(`{"username": "sh", "role": "student"}`),
	}

	database.AddGroupWithMembers(database.Group{
		Title:    "smth 2",
		Students: []string{"sh"},
	},
		members)
}

func (h *Handler) home(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))

	if err != nil {
		// TODO return error
		return
	}
	switch c.GetString("role") {
	case "student":
		{
			c.HTML(http.StatusOK, "home.tmpl", gin.H{"Name3": profile.Name3})
		}
	case "admin":
		{
			c.HTML(http.StatusOK, "home-admin.tmpl", gin.H{"Name3": profile.Name3})
		}
	}

}

func (h *Handler) addContest(c *gin.Context) {
	title := "Мой первый контест"
	access := map[string]interface{}{"public": true}
	groupOwner := 1
	startTime := time.Now()
	duration := time.Hour

	// Вызов функции для создания контеста
	contestID, err := database.CreateContest(title, access, groupOwner, startTime, duration)
	if err != nil {
		fmt.Println("Ошибка при создании контеста:", err)
		return
	}

	fmt.Printf("Контест создан с ID: %d\n", contestID)

	c.HTML(http.StatusOK, "home.tmpl", gin.H{})
}
