package handler

import (
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

type message struct {
	Task   string `json:"task"`
	Sender string `json:"sender"`
	Git    string `json:"git"`
	Time   string `json:"time"`
}

type msgtheme struct {
	Theme string    `json:"theme"`
	Msgs  []message `json:"msgs"`
}

type arrtheme []msgtheme

func (a arrtheme) Len() int      { return len(a) }
func (a arrtheme) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a arrtheme) Less(i, j int) bool {
	a1, a2 := a[i].Msgs[0].Time, a[j].Msgs[0].Time
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

func (h *Handler) forumMainPage(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))
	if err != nil {
		// TODO return error
		return
	}
	//s, _ := database.GetAllUserThemes(c.GetString("username"))
	res := make([]msgtheme, 0)
	_, themes, tasks, senders, ti, e := database.GetUserForum(c.GetString("username"))
	if e != nil {
		fmt.Println(e)
		return
	}
	if len(themes) != 0 {
		res = make([]msgtheme, 1)
		le := 0
		git, fi, la := database.GetInfoForForumProfilePage(senders[0])
		q := fi + " " + la
		giturl := "https://github.com/" + git
		res[0] = msgtheme{
			Theme: themes[0],
			Msgs: []message{message{
				Task:   tasks[0],
				Sender: q,
				Git:    giturl,
				Time:   ti[0],
			}},
		}

		for i := 1; i < len(themes); i++ {
			git, fi, la := database.GetInfoForForumProfilePage(senders[i])
			q := fi + " " + la
			giturl := "https://github.com/" + git
			if themes[i-1] == themes[i] {
				res[le].Msgs = append(res[le].Msgs, message{
					Task:   tasks[i],
					Sender: q,
					Git:    giturl,
					Time:   ti[i],
				})
			} else {
				le++
				res = append(res, msgtheme{
					Theme: themes[i],
					Msgs: []message{message{
						Task:   tasks[i],
						Sender: q,
						Git:    giturl,
						Time:   ti[i],
					}},
				})
			}
		}
	}
	for _, thems := range res {
		msgs := thems.Msgs
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
	}
	sort.Sort(arrtheme(res))
	//fmt.Println(k)
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
