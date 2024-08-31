package handler

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

type setuserpoints []rattable

func (a setuserpoints) Len() int           { return len(a) }
func (a setuserpoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a setuserpoints) Less(i, j int) bool { return a[i].Points > a[j].Points }

type rattable struct {
	Username string `json:"username"`
	Points   int    `json:"points"`
	Place    int    `json:"place"`
}

func (h *Handler) rating(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))

	if err != nil {
		// TODO return error
		return
	}

	rank, err := database.GetRankInfo()
	if err != nil {
		// TODO return error
		return
	}
	res := make([]rattable, 0)
	for k, v := range rank {
		var res1 rattable
		u, err := database.GetInfoForProfilePage(k)
		if err != nil {
			fmt.Println(err)
			return
		}
		fila := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		res1.Username, res1.Points = fila, v
		res = append(res, res1)
	}
	sort.Sort(setuserpoints(res))
	for i := 0; i < len(res); i++ {
		res[i].Place = i + 1
	}
	c.HTML(http.StatusOK, "rating.tmpl", gin.H{
		"Name3": profile.Name3,
		"list1": res,
	})
}
