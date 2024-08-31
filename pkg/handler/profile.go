package handler

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

type arruserpoints []userswithpoints

func (a arruserpoints) Len() int           { return len(a) }
func (a arruserpoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a arruserpoints) Less(i, j int) bool { return a[i].points > a[j].points }

type userswithpoints struct {
	user   string
	points int
}

func getuserplace(user string) (int, error) {
	res := make([]userswithpoints, 0)
	RankInfo, err := database.GetRankInfo()
	if err != nil {
		return 0, err
	}
	for k, v := range RankInfo {
		var a userswithpoints
		a.user = k
		a.points = v
		res = append(res, a)
	}
	sort.Sort(arruserpoints(res))
	for i := 0; i < len(res); i++ {
		if res[i].user == user {
			return i + 1, nil
		}
	}
	return 0, nil
}

func (h *Handler) profileMainPage(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))
	if err != nil {
		// TODO return error
		return
	}
	w, e := database.GetUserSols(c.GetString("username"))
	fmt.Println(w, e)
	rank, err := getuserplace(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rank)
	a, _ := database.GetRankInfo()
	fmt.Println(a)
	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"NickName": profile.Username,
		"Surname":  profile.LastName,
		"Name":     profile.FirstName,
		"Name3":    profile.Name3,
		"Email":    profile.Email,
		"rank":     rank,
		//"Avatar":   profile.Avatar,
	})
}
