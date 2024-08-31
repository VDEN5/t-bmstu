package handler

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	//"time"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

type letform struct {
	//TODO: наверное стоит самому время устанавливать
}

type lettasks struct {
	Task   string `json:"task"`
	Number int    `json:"number"`
}

type letres struct {
	Task   string `json:"task"`
	Number int    `json:"number"`
	Res    int    `json:"res"`
}

func (h *Handler) letuchka(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))
	if err != nil {
		fmt.Println(err)
		return
	}
	w, e := database.ExistLet(c.GetString("username"), 1)
	if e != nil {
		fmt.Println(err)
		return
	}
	w1, e := database.ExistLet(c.GetString("username"), 2)
	if e != nil {
		fmt.Println(err)
		return
	}
	if w {
		let, err := database.GetLet(c.GetString("username"), 1)
		if err != nil {
			fmt.Println(err)
			return
		}
		if let.Starttime.Add(time.Hour * 4).After(time.Now().Add(time.Hour * 3)) { //<
			//идет сейчас, показать задания
			res := make([]lettasks, 0)
			for i, t := range let.Tasks {
				s := "/view/problem/" + base64.StdEncoding.EncodeToString([]byte(t))
				fmt.Println(s)
				res = append(res, lettasks{
					Number: i + 1,
					Task:   s,
				})
			}
			c.HTML(http.StatusOK, "letuchka.tmpl", gin.H{
				"Name3": profile.Name3,
				"list":  res,
			})
			fmt.Println(let)
			return
		} else {
			//требует утверждения результатов, показ утверждения резов, но до этого перевести состояние в 2 и перенести все решенные
			let.Status = 2
			database.Updstattofin(let)
			for i, t := range let.Tasks {
				if t[0] == 'a' {
					ex, err := database.ExistSols(c.GetString("username"), t[0:4], t[4:])
					if err != nil {
						fmt.Println(err)
						return
					}
					if ex {
						s, e := database.GetSol(c.GetString("username"), t[0:4], t[4:])
						if e != nil {
							fmt.Println(e)
							return
						}
						if s.Time.Before(let.Starttime.Add(time.Hour * 4)) {
							let.Soltasks = append(let.Soltasks, strconv.Itoa(i+1))
						}
					}
				} else if t[0] == 't' {
					ex, err := database.ExistSols(c.GetString("username"), t[0:5], t[5:])
					if err != nil {
						fmt.Println(err)
						return
					}
					if ex {
						s, e := database.GetSol(c.GetString("username"), t[0:5], t[5:])
						if e != nil {
							fmt.Println(e)
							return
						}
						if s.Time.Before(let.Starttime.Add(time.Hour * 4)) {
							let.Soltasks = append(let.Soltasks, strconv.Itoa(i+1))
						}
					}
				}
			}
			fmt.Println(let.Soltasks)
			database.Updsols(let)
			fmt.Println(let)
			Letres := make(map[int]letres)
			for i, t := range let.Tasks {
				s := "/view/problem/" + base64.StdEncoding.EncodeToString([]byte(t))
				Letres[i+1] = letres{
					Task:   s,
					Number: i + 1,
					Res:    0,
				}
			}
			for _, strt := range let.Soltasks {
				t, _ := strconv.Atoi(strt)
				tmp := Letres[t]
				Letres[t] = letres{
					Task:   tmp.Task,
					Number: tmp.Number,
					Res:    1,
				}
			}
			requestMethod := c.Request.Method
			switch requestMethod {
			case "GET":
				{
					c.HTML(http.StatusOK, "letuchkabutton.tmpl", gin.H{
						"Name3":   profile.Name3,
						"Buttext": "Подтверждаю результаты",
						"letres":  Letres,
					})
				}
			case "POST":
				{
					var form letform
					if err := c.BindJSON(&form); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
						fmt.Println("achtung")
						return
					}
					let.Status = 3
					database.Updstattofin(let)
				}
			default:
				{
					c.JSON(http.StatusBadRequest, "No such router for this method")
				}
			}
			fmt.Println("vkbkjvbk")
			return
		}
	} else if w1 {
		//требует утверждения результатов, показ утверждения резов
		let, err := database.GetLet(c.GetString("username"), 2)
		if err != nil {
			fmt.Println(err)
			return
		}
		Letres := make(map[int]letres)
		for i, t := range let.Tasks {
			s := "/view/problem/" + base64.StdEncoding.EncodeToString([]byte(t))
			Letres[i+1] = letres{
				Task:   s,
				Number: i + 1,
				Res:    0,
			}
		}
		for _, strt := range let.Soltasks {
			t, _ := strconv.Atoi(strt)
			tmp := Letres[t]
			Letres[t] = letres{
				Task:   tmp.Task,
				Number: tmp.Number,
				Res:    1,
			}
		}
		requestMethod := c.Request.Method
		switch requestMethod {
		case "GET":
			{
				c.HTML(http.StatusOK, "letuchkabutton.tmpl", gin.H{
					"Name3":   profile.Name3,
					"Buttext": "Подтверждаю результаты",
					"letres":  Letres,
				})
			}
		case "POST":
			{
				var form letform
				if err := c.BindJSON(&form); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
					fmt.Println("achtung")
					return
				}
				let, err := database.GetLet(c.GetString("username"), 2)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(let)
				let.Status = 3
				database.Updstattofin(let)
				fmt.Println(let)
			}
		default:
			{
				c.JSON(http.StatusBadRequest, "No such router for this method")
			}
		}
		fmt.Println("dlcpeork")
		return
	} else {
		//не надо ничего утверждать и летучки нет, показать создание летучки
		requestMethod := c.Request.Method
		switch requestMethod {
		case "GET":
			{
				c.HTML(http.StatusOK, "letuchkabutton.tmpl", gin.H{
					"Name3":   profile.Name3,
					"Buttext": "Создать новую летучку"})
			}
		case "POST":
			{
				var form letform
				if err := c.BindJSON(&form); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
					fmt.Println("achtung")
					return
				}
				gentask := make([]string, 0)
				for len(gentask) < 5 { //5 problems
					testsys := rand.Intn(3-1) + 1 //1-acmp,2-timus
					if testsys == 1 {
						prid := strconv.Itoa(rand.Intn(100-1) + 1)
						w, e := database.ExistSols(c.GetString("username"), "acmp", prid)
						if e != nil {
							fmt.Println(e)
							return
						}
						if !w {
							gentask = append(gentask, "acmp"+prid)
						}
					}
					if testsys == 2 {
						prid := strconv.Itoa(rand.Intn(1100-1000) + 1000)
						w, e := database.ExistSols(c.GetString("username"), "timus", prid)
						if e != nil {
							fmt.Println(e)
							return
						}
						if !w {
							gentask = append(gentask, "timus"+prid)
						}
					}
				}
				database.CreateLet(database.Letuchka{
					Tasks:     gentask,
					Soltasks:  []string{},
					Letuser:   c.GetString("username"),
					Starttime: time.Now(),
					Status:    1,
				})
			}
		default:
			{
				c.JSON(http.StatusBadRequest, "No such router for this method")
			}
		}
		return
	}
}
