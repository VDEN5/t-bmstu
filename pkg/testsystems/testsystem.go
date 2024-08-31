package testsystems

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/acmp"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/codeforces"
	"github.com/VDEN5/t-bmstu/pkg/testsystems/timus"
	"github.com/VDEN5/t-bmstu/pkg/websockets"
)

func getpointstimus(id int) int {
	reqURL := "https://timus.online/problem.aspx?space=1&num=" + strconv.Itoa(id)
	resp, err := http.Get(reqURL)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	str := string(body)
	ind := strings.Index(str, "problem_links")
	str = str[ind+33:]
	ind = strings.Index(str, "</SPAN>")
	str = str[:ind]
	res, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("this is not number %s", err)
		return -1
	}
	return res
}

func getpointsacmp(id int) int {
	reqURL := "https://acmp.ru/index.asp?main=task&id_task=" + strconv.Itoa(id)
	resp, err := http.Get(reqURL)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	str := string(body)
	ind := strings.Index(str, "</h1>")
	str = str[ind:]
	ind = strings.Index(str, "</i>")
	str = str[:ind]
	ind = strings.Index(str, "%")
	str = str[ind-8 : ind]
	ind = strings.Index(str, ": ")
	str = str[ind+2:]
	res, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("this is not number %s", err)
		return -1
	}
	return res
}

// AllowedTestsystems - разрешенные (добавленные) тестирующие системы
var AllowedTestsystems = []TestSystem{
	&timus.Timus{Name: "timus"},
	&acmp.ACMP{Name: "acmp"},
	&codeforces.Codeforces{Name: "codeforces"},
}

// TestSystem - это интерфейс класса тестирующей системы, то есть все тестирующие системы должны обладать этими функциями
type TestSystem interface {
	Init()
	// GetName - получить имя тестирущей системы
	GetName() string
	// CheckLanguage - проверяет, существует ли у данной тестирующей системы такой язык программирования
	CheckLanguage(language string) bool
	// GetLanguages - получить языки на которых можно сдавать в этой тестирующей системе
	GetLanguages() []string
	// Submitter - воркер, который занимается отправлением посылок, и будет запускаться в отдельной горутине
	Submitter(wg *sync.WaitGroup, ch chan<- database.Submission)
	// GetProblem - получить условие задачи !!! ALERT, его надо получать по частям, см -> database.Task
	GetProblem(taskID string) (database.Task, error)
	// Checker - воркер, который занимается обновлением статусов посылок
	Checker(wg *sync.WaitGroup, ch chan<- database.Submission)
}

var wg sync.WaitGroup

func InitGorutines() error {
	submitterChannels := make(map[string]chan database.Submission)
	checkerChannels := make(map[string]chan database.Submission)

	for _, TestSystem := range AllowedTestsystems {
		ch1 := make(chan database.Submission)
		submitterChannels[TestSystem.GetName()] = ch1

		ch2 := make(chan database.Submission)
		checkerChannels[TestSystem.GetName()] = ch2

		// инициализация
		TestSystem.Init()

		wg.Add(2)

		go TestSystem.Submitter(&wg, ch1)
		go TestSystem.Checker(&wg, ch2)
	}

	// запустим горутины для самбиттеров
	for _, ch := range submitterChannels {
		go func(c <-chan database.Submission) {
			for msg := range c {
				// надо обновить запись в базе данных
				err := database.UpdateSubmissionData(msg)
				if err != nil {
					fmt.Println(err)
				}

				// передать по веб-сокету
				go websockets.SendMessageToUser(msg.SenderLogin, msg)
			}
		}(ch)
	}

	// для чекеров
	for _, ch := range checkerChannels {
		go func(c <-chan database.Submission) {
			for msg := range c {
				// надо обновить запись в базе данных
				err := database.UpdateSubmissionData(msg)
				if err != nil {
					fmt.Println(err)
				}

				// если проверка была окончена, то записать это в соответствующем поле у пользователя и в контест
				if msg.Status == 2 {

				}

				// передать по веб-сокету
				go websockets.SendMessageToUser(msg.SenderLogin, msg)
				if msg.Verdict == "Accepted" {
					fmt.Println("wow")
					ex, err := database.ExistSols(msg.SenderLogin, msg.TestingSystem, msg.TaskID)
					if err != nil {
						fmt.Println(err)
					}

					if !ex {
						a, err := strconv.Atoi(msg.TaskID)
						if err != nil {
							fmt.Println(err)
						}

						b := 0
						if msg.TestingSystem == "acmp" {
							b = getpointsacmp(a)
						} else {
							if msg.TestingSystem == "timus" {
								b = getpointstimus(a)
							}
						}
						database.CreateSolution(database.Ranktable{
							RankUser:   msg.SenderLogin,
							TestSystem: msg.TestingSystem,
							ProblemId:  a,
							Points:     b,
							Time:       time.Now(),
						})
					}
				}
			}
		}(ch)
	}

	return nil
}
