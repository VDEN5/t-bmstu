package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"sort"
	"strconv"
	"time"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

type forumform struct {
	Theme string `json:"username1"`
	Task1 string `json:"password1"`
	//Avatar    string `json:"avatar"`
}

type themeform struct {
	Task1 string `json:"password1"`
	//Avatar    string `json:"avatar"`
}

type message struct {
	Task   string `json:"task"`
	Sender string `json:"sender"`
	Git    string `json:"git"`
	Time   string `json:"time"`
}

type msgtheme struct {
	Theme   string    `json:"theme"`
	Theme1  string    `json:"theme1"`
	Msgs    []message `json:"msgs"`
	User    string    `json:"user"`
	Time    string    `json:"time"`
	Task    string    `json:"task"`
	Userava string    `json:"userava"`
}

type arrtheme []msgtheme

func (a arrtheme) Len() int      { return len(a) }
func (a arrtheme) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a arrtheme) Less(i, j int) bool {
	a1, a2 := a[i].Time, a[j].Time
	layout := "2006-01-02 15:04:05"
	b1, err := time.Parse(layout, a1)
	if err != nil {
		fmt.Println("Error parsing time string:", err)
		return false
	}
	b2, err := time.Parse(layout, a2)
	if err != nil {
		fmt.Println("Error parsing time string:", err)
		return false
	}
	return b1.After(b2)
}

type ft struct {
	Task   string `json:"task"`
	Sender string `json:"sender"`
	Time   string `json:"time"`
	Git    string `json:"git"`
	Ava    string `json:"ava"`
	You    bool   `json:"you"`
}

func (h *Handler) getmsgid(c *gin.Context) {
	stringSubmissionId := c.Param("id")
	submissionId, err := strconv.Atoi(stringSubmissionId)
	if err != nil {
		c.String(404, "There are no such submission")
		return
	}
	code1, err := database.GetTheme(submissionId)
	code, err := database.GetTask(submissionId)
	//_, code, err := database.MessageExist(submissionId)
	if err != nil {
		c.String(404, "There are no such submission")
		return
	}
	c.String(200, code1+"<br>"+code)
}

func (h *Handler) getforumtheme(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))
	decoded1, err := base64.StdEncoding.DecodeString(c.Param("id"))
	if err != nil {
		fmt.Println("Error decoding:", err)
		return
	}
	stringSubmissionId := string(decoded1)
	if err != nil {
		// TODO return error
		return
	}
	ta, se, ti, e := database.GetTasksFromTheme1(stringSubmissionId)
	if e != nil {
		fmt.Println(e)
		return
	}
	res := make([]ft, 0)
	for i := 0; i < len(se); i++ {
		git, fi, la := database.GetInfoForForumProfilePage(se[i])
		q := fi + " " + la
		giturl := "https://github.com/" + git
		if err != nil {
			fmt.Println(err)
			return
		}
		profileforum, err := database.GetInfoForProfilePage(se[i])
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, ft{
			Task:   ta[i],
			Sender: q,
			Time:   ti[i],
			Git:    giturl,
			You:    se[i] == c.GetString("username"),
			Ava:    profileforum.Name3,
		})
	}

	//fmt.Println(res, e)

	// c.HTML(http.StatusOK, "theme.tmpl", gin.H{
	// 	"themelist": res,
	// 	"Name3":     profile.Name3,
	// })
	requestMethod := c.Request.Method
	switch requestMethod {
	case "GET":
		{
			c.HTML(http.StatusOK, "theme.tmpl", gin.H{"Name3": profile.Name3, "themelist": res})
		}
	case "POST":
		{
			var form themeform
			if err := c.BindJSON(&form); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
				fmt.Println("achtung")
				return
			}
			currentTime := time.Now()
			strtime := currentTime.Format("2006-01-02 15:04:05")
			database.CreateMessage(database.Message{
				ForumUser:  c.GetString("username"),
				ForumTheme: stringSubmissionId,
				ForumTask:  form.Task1,
				MSGtime:    strtime,
			})

			//c.Redirect(302, "/view/forum")
		}
	default:
		{
			c.JSON(http.StatusBadRequest, "No such router for this method")
		}
	}
}

func (h *Handler) forumMainPage(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	res1, err := database.GetAllUserThemes(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	res := make([]msgtheme, 0)
	for th, me := range res1 {
		_, fi, la := database.GetInfoForForumProfilePage(me.ForumUser)
		q := fi + " " + la
		profileforum, err := database.GetInfoForProfilePage(me.ForumUser)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, msgtheme{
			Theme:   th,
			Theme1:  base64.StdEncoding.EncodeToString([]byte(th)),
			User:    q,
			Time:    me.MSGtime,
			Task:    me.ForumTask,
			Userava: profileforum.Name3,
		})
	}
	sort.Sort(arrtheme(res))
	requestMethod := c.Request.Method
	switch requestMethod {
	case "GET":
		{
			c.HTML(http.StatusOK, "forum.tmpl", gin.H{"Name3": profile.Name3, "list": res})
		}
	case "POST":
		{
			var form forumform
			if err := c.BindJSON(&form); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
				fmt.Println("achtung")
				return
			}
			currentTime := time.Now()
			strtime := currentTime.Format("2006-01-02 15:04:05")
			database.CreateMessage(database.Message{
				ForumUser:  c.GetString("username"),
				ForumTheme: form.Theme,
				ForumTask:  form.Task1,
				MSGtime:    strtime,
			})

			//c.Redirect(302, "/view/forum")
		}
	default:
		{
			c.JSON(http.StatusBadRequest, "No such router for this method")
		}
	}
	//c.HTML(http.StatusOK, "forum.tmpl", gin.H{"Name3": profile.Name3})
}
