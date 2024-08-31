package codeforces

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/VDEN5/t-bmstu/pkg/database"
	"golang.org/x/net/html"
)

type Codeforces struct {
	Name string
}

func (t *Codeforces) Init() {

}

func (t *Codeforces) GetName() string {
	return t.Name
}

func (t *Codeforces) CheckLanguage(language string) bool {
	languagesDict := map[string]struct{}{
		"GNU GCC C11 5.1.0":                struct{}{},
		"GNU G++ 14 6.4.0":                 struct{}{},
		"GNU G++ 17 7.3.0":                 struct{}{},
		"GNU G++20 13.2 (64 bit, winlibs)": struct{}{},
		"C# 8, .NET Core 3.1":              struct{}{},
		"C# 10, .NET SDK 6.0":              struct{}{},
		"C# Mono 6.8":                      struct{}{},
		"D DMD32 v2.105.0":                 struct{}{},
		"Python 3.8.10":                    struct{}{},
	}

	_, exist := languagesDict[language]

	if !exist {
		return false
	}

	return true
}

func GetBody(client *http.Client, URL string) ([]byte, error) {
	resp, err := client.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func findCsrf(body []byte) (string, error) {
	reg := regexp.MustCompile(`csrf='(.+?)'`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New("Cannot find csrf")
	}
	return string(tmp[1]), nil
}

func (t *Codeforces) GetLanguages() []string {
	return []string{
		"GNU GCC C11 5.1.0",
		"GNU G++ 14 6.4.0",
		"GNU G++ 17 7.3.0",
		"GNU G++20 13.2 (64 bit, winlibs)",
		"C# 8, .NET Core 3.1",
		"C# 10, .NET SDK 6.0",
		"C# Mono 6.8",
		"D DMD32 v2.105.0",
		"Python 3.8.10",
	}
}

func (t *Codeforces) GetProblem(taskID string) (database.Task, error) {
	taskURL := fmt.Sprintf("https://codeforces.com/problemset/problem/%s/?locale=ru", taskID)
	req, err := http.NewRequest("GET", taskURL, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return database.Task{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return database.Task{}, err
	}

	taskName := ""

	doc.Find("div.title").Each(func(i int, s *goquery.Selection) {
		if taskName == "" {
			index := strings.IndexByte(s.Text(), '.')
			taskName = s.Text()[index+1:]
		}

	})

	Constraints := map[string]string{}

	var Condition string

	doc.Find("div.header").NextUntil("div.input-specification").Each(func(i int, s *goquery.Selection) {
		Condition = s.Text()
	})
	Input := ""
	Output := ""

	//Входные данные
	doc.Find("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.input-specification div.section-title").NextUntil("div.input-specification").Each(func(i int, s *goquery.Selection) {
		Input = s.Text()
	})

	//Выходные данные
	doc.Find("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.output-specification div.section-title").NextUntil("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests").Each(func(i int, s *goquery.Selection) {
		Output = s.Text()
	})

	if err != nil {
		log.Fatal(err)
		return database.Task{}, err

	}

	//Тесты(на codeforces они есть не всегда)
	tests := parseTableToJSON(doc.Find("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests"), doc)

	return database.Task{
		Name:        taskName,
		Condition:   Condition,
		Constraints: Constraints,
		InputData:   Input,
		OutputData:  Output,
		Tests: map[string]interface{}{
			"tests": tests,
		},
	}, nil
}

func (t *Codeforces) Submitter(wg *sync.WaitGroup, ch chan<- database.Submission) {

	// Создаем новый cookie jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Выполняем первый запрос для получения CSRF-токена
	//Это токен, который каждый раз новый и поэтому его надо запарсить
	resp, err := client.Get("https://codeforces.com/enter?back=%2F")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	doc, err := html.Parse(bytes.NewReader(htmlData))
	if err != nil {
		fmt.Println(err)
	}

	var csrfToken string
	var findCSRFToken func(*html.Node)
	findCSRFToken = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "csrf_token" {
					for _, a := range n.Attr {
						if a.Key == "value" {
							csrfToken = a.Val
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCSRFToken(c)
		}
	}
	findCSRFToken(doc)

	// Теперь, когда у нас есть CSRF-токен, мы можем использовать его в последующих запросах

	loginURL := "https://Codeforces.com/enter?back=%2F"
	loginData := url.Values{
		"csrf_token":    {csrfToken},
		"action":        {"enter"},
		"ftaa":          {"nxnnepf797p929r93p"},
		"bfaa":          {"939f6b320d3e9e423cd3b4899db9631d"},
		"handleOrEmail": {"gordejka179"},
		"password":      {"XB#8T^m;xj5n~;8"},
		"_tta":          {"35"},
	}

	resp, err = client.PostForm(loginURL, loginData)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	myToCodeforcesDict := map[string]string{
		"GNU GCC C11 5.1.0":                "43",
		"GNU G++ 14 6.4.0":                 "50",
		"GNU G++ 17 7.3.0":                 "54",
		"GNU G++20 13.2 (64 bit, winlibs)": "89",
		"C# 8, .NET Core 3.1":              "65",
		"C# 10, .NET SDK 6.0":              "79",
		"C# Mono 6.8":                      "9",
		"D DMD32 v2.105.0":                 "28",
		"Python 3.8.10":                    "31",
	}

	for {
		submissions, err := database.GetSubmitsWithStatus(t.GetName(), 0)
		if err != nil {
			fmt.Println(err)
		}

		// перебираем все решения
		for _, submission := range submissions {

			body, err := GetBody(client, fmt.Sprintf("https://codeforces.com/problemset/submit/%s", submission.TaskID))
			if err != nil {
				return
			}

			csrf, err := findCsrf(body)
			if err != nil {
				return
			}

			fileData := url.Values{
				"csrf_token": {csrf},
				"ftaa":       {"nxnnepf797p929r93p"},
				"bfaa":       {"939f6b320d3e9e423cd3b4899db9631d"},
				"action":     {"submitSolutionFormSubmitted"},

				"contestId":             {string(strings.Split(submission.TaskID, "/")[0])},
				"submittedProblemIndex": {string(strings.Split(submission.TaskID, "/")[1])},

				"programTypeId": {myToCodeforcesDict[submission.Language]},
				"source":        {string(submission.Code)},
				"tabSize":       {"4"},
				"sourceFile":    {""},
				"_tta":          {"35"},
			}

			submitURL := fmt.Sprintf("https://codeforces.com/problemset/submit/%s", submission.TaskID)
			resp, err = client.PostForm(submitURL, fileData)
			//Теперь нас перенесло на таблицу с посылками

			doc, err := goquery.NewDocumentFromReader(resp.Body)

			//ищем id нашей посылки
			var id string

			table := doc.Find("table")
			rows := table.Find("tr")
			row := rows.Eq(1)
			columns := row.Find("td")
			columns.Each(func(j int, cell *goquery.Selection) {
				if j == 0 {
					id = strings.TrimSpace(cell.Text())
				}
			})

			// теперь надо передать по каналу, что был изменен статус этой задачи
			submission.Status = 1
			submission.Verdict = "Compiling"
			submission.SubmissionNumber = id
			ch <- submission

		}

		time.Sleep(time.Second * 2)
	}

}

func (t *Codeforces) Checker(wg *sync.WaitGroup, ch chan<- database.Submission) {
	for {
		submissions, err := database.GetSubmitsWithStatus(t.GetName(), 1)

		if err != nil {
			fmt.Println(err)
		}

		submissionsDict := make(map[string]database.Submission)
		submissionsIDs := make([]string, 0)

		type UserStatus struct {
			Status string `json:"status"`
			Result []struct {
				ID        int `json:"id"`
				ContestID int `json:"contestId"`
				Problem   struct {
					Index string `json:"index"`
					Name  string `json:"name"`
					Type  string `json:"type"`
				} `json:"problem"`
				Verdict             string `json:"verdict"`
				TimeConsumedMillis  int    `json:"timeConsumedMillis"`
				MemoryConsumedBytes int    `json:"memoryConsumedBytes"`
			} `json:"result"`
		}

		for _, submission := range submissions {
			submissionsDict[submission.SubmissionNumber] = submission
			submissionsIDs = append(submissionsIDs, submission.SubmissionNumber)
		}

		for len(submissions) != 0 {

			resp, err := http.Get("https://codeforces.com/api/user.status?handle=gordejka179&from=1&count=10")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			var userStatus UserStatus
			err = json.Unmarshal(body, &userStatus)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Выводим информацию о попытках решения
			for _, attempt := range userStatus.Result {
				idStr := strconv.Itoa(attempt.ID)
				for _, submissionID := range submissionsIDs {
					if idStr == submissionID {
						submission, exists := submissionsDict[idStr]
						if !exists {
							log.Println("не найдена попытка:", idStr)
							return
						}
						delete(submissionsDict, idStr)

						for i, id := range submissionsIDs {
							if id == idStr {
								submissionsIDs = append(submissionsIDs[:i], submissionsIDs[i+1:]...)
								break
							}
						}

						submissions = submissions[1:]

						verdict := attempt.Verdict
						test := "-"
						executionTime := strconv.Itoa(attempt.TimeConsumedMillis)
						memoryUsed := strconv.Itoa(attempt.MemoryConsumedBytes)

						submission.Verdict = verdict
						submission.Test = test
						submission.ExecutionTime = executionTime
						submission.MemoryUsed = memoryUsed

						if endChecking(verdict) {
							submission.Status = 2
						}
						ch <- submission

					}
				}

			}
		}

		time.Sleep(time.Second * 2)
	}

}
