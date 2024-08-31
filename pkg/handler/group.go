package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createGroup(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		{
			c.HTML(http.StatusOK, "create-group.tmpl", gin.H{})
		}
	case "POST":
		{
			var requestData struct {
				GroupName  string `json:"groupName"`
				InviteLink string `json:"inviteLink"`
			}

			if err := c.ShouldBindJSON(&requestData); err != nil {
				c.Status(http.StatusBadRequest)
				return
			}

			exist, err := database.CheckGroupExist(requestData.GroupName, requestData.InviteLink)
			if err != nil {
				c.Status(http.StatusBadRequest)
				return
			}

			if exist {
				c.JSON(http.StatusBadRequest, gin.H{"error": "such group exists"})
				return
			}

			database.AddGroupWithMembers(database.Group{
				Title:      requestData.GroupName,
				InviteCode: []byte(requestData.InviteLink),
			},
				[]json.RawMessage{})

			c.Status(http.StatusOK)
		}
	}
}

func (h *Handler) getGroupContests(c *gin.Context) {
	profile, err1 := database.GetInfoForProfilePage(c.GetString("username"))

	if err1 != nil {
		// TODO return error
		return
	}
	groupId, err := strconv.Atoi(c.Param("group_id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	contests, err := database.GetGroupContests(groupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	type Contest struct {
		Title    string `json:"title"`
		ID       int
		TimeLeft string `json:"timeleft"`
	}

	// костыль
	currentTime := time.Now().Add(3 * time.Hour)
	var s string
	contestsForTemplate := []Contest{}
	for _, contest := range contests {
		endTime := contest.StartTime.Add(contest.Duration).In(currentTime.Location())
		var timeRemaining time.Duration
		if currentTime.Before(endTime) {
			timeRemaining = endTime.Sub(currentTime)
			hours := int(timeRemaining.Hours())
			minutes := int(timeRemaining.Minutes()) % 60
			seconds := int(timeRemaining.Seconds()) % 60
			s = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		} else {
			s = ""
		}
		fmt.Println(s)
		contestsForTemplate = append(contestsForTemplate, Contest{
			Title:    contest.Title,
			ID:       contest.ID,
			TimeLeft: s,
		})
	}

	c.HTML(http.StatusOK, "group_contests.tmpl", gin.H{
		"Contests": contestsForTemplate,
		"Name3":    profile.Name3,
	})
}

func (h *Handler) checkInvite(c *gin.Context) {
	inviteHash := c.Param("invite_hash")

	if inviteHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no such group"})
		return
	}

	exist, groupId, err := database.CheckInviteCode(inviteHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no such group"})
		return
	}

	database.AddUserToGroup(c.GetString("username"), groupId, "student")
	c.JSON(http.StatusOK, gin.H{"Success": "U are member of this group now"})
}
