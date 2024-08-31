package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/acmp"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/codeforces"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/timus"
	"github.com/gin-gonic/gin"
)

func (h *Handler) acmpTaskList(c *gin.Context) {
	count := c.Query("count")

	parsedCount, err := strconv.Atoi(count)
	if err != nil {
		parsedCount = 15
	}

	if parsedCount > 50 {
		parsedCount = 50
	}

	usersols, err := database.GetUserSols(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	mem := make(map[string]bool)
	for _, k := range usersols {
		if k.TestSystem == "acmp" {
			mem[strconv.Itoa(k.ProblemId)] = true
		}
	}
	taskList, err := acmp.GetTaskList(parsedCount, c.GetString("username"))
	for i, g := range taskList {
		asd, sdf := base64.StdEncoding.DecodeString(g.ID)
		id1 := string(asd)[4:]
		if sdf != nil {
			fmt.Println(err)
			return
		}
		if mem[id1] {
			num, err := strconv.Atoi(g.Points)
			if err != nil {
				fmt.Println(err)
				return
			}
			taskList[i].Yourpoi = num
		}
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad req")
	}

	profile, err1 := database.GetInfoForProfilePage(c.GetString("username"))

	if err1 != nil {
		// TODO return error
		return
	}

	c.HTML(http.StatusOK, "testsystem-tasks-list.tmpl", gin.H{
		"TestSystem": "ACMP",
		"Tasks":      taskList,
		"Name3":      profile.Name3,
	})
}
func (h *Handler) timusTaskList(c *gin.Context) {
	count := c.Query("count")

	parsedCount, err := strconv.Atoi(count)
	if err != nil {
		parsedCount = 15
	}

	if parsedCount > 50 || parsedCount <= 0 {
		parsedCount = 20
	}

	from := c.Query("from")

	parsedFrom, err := strconv.Atoi(from)
	if err != nil {
		parsedFrom = 1
	}

	if parsedFrom <= 0 {
		parsedFrom = 1
	}

	usersols, err := database.GetUserSols(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	mem := make(map[string]bool)
	for _, k := range usersols {
		if k.TestSystem == "timus" {
			mem[strconv.Itoa(k.ProblemId)] = true
		}
	}
	taskList, err := timus.GetTaskList(parsedFrom, parsedCount, c.GetString("username"))
	for i, g := range taskList {
		asd, sdf := base64.StdEncoding.DecodeString(g.ID)
		id1 := string(asd)[5:]
		if sdf != nil {
			fmt.Println(err)
			return
		}
		if mem[id1] {
			num, err := strconv.Atoi(g.Points)
			if err != nil {
				fmt.Println(err)
				return
			}
			taskList[i].Yourpoi = num
		}
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad req")
	}
	profile, err1 := database.GetInfoForProfilePage(c.GetString("username"))

	if err1 != nil {
		// TODO return error
		return
	}

	c.HTML(http.StatusOK, "testsystem-tasks-list.tmpl", gin.H{
		"TestSystem": "Timus",
		"Tasks":      taskList,
		"Name3":      profile.Name3,
	})
}

func (h *Handler) codeforcesTaskList(c *gin.Context) {

	count := c.Query("count")

	parsedCount, err := strconv.Atoi(count)
	if err != nil {
		parsedCount = 15
	}

	if parsedCount > 50 {
		parsedCount = 50
	}

	from := c.Query("from")

	parsedFrom, err := strconv.Atoi(from)
	if err != nil {
		parsedFrom = 1
	}

	if parsedFrom <= 0 {
		parsedFrom = 1
	}

	taskList, err := codeforces.GetTaskList(parsedFrom, parsedCount)
	for i, _ := range taskList {
		taskList[i].Points = "0"
		taskList[i].Yourpoi = 0
	} //TODO: create codeforces points!!!
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad req")
	}
	profile, err1 := database.GetInfoForProfilePage(c.GetString("username"))

	if err1 != nil {
		// TODO return error
		return
	}

	c.HTML(http.StatusOK, "testsystem-tasks-list.tmpl", gin.H{
		"TestSystem": "codeforces",
		"Tasks":      taskList,
		"Name3":      profile.Name3,
	})

}

func (h *Handler) submitTask(c *gin.Context) {
	var requestData struct {
		SourceCode string `json:"sourceCode"`
		Language   string `json:"language"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestData.Language == "Select language " {
		c.JSON(http.StatusBadRequest, "There are not such language")
		return
	}

	err := TaskSubmit(c.Param("id"), c.GetString("username"), requestData.SourceCode, requestData.Language,
		-1, -1)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task submitted successfully",
	})
}
