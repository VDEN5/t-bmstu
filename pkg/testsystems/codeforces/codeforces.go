package codeforces

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gordejka179/t-bmstu/pkg/database"
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
		"GNU GCC C11 5.1.0":           struct{}{},
		"GNU G++ 14 6.4.0":            struct{}{},
		"Python 3.8.10":               struct{}{},
		"PascalABC.NET 3.8.3":         struct{}{},
		"Java SE JDK 16.0.1":          struct{}{},
		"Free Pascal 3.2.2":           struct{}{},
		"Borland Delphi 7.0":          struct{}{},
		"Microsoft Visual C++ 2017":   struct{}{},
		"Microsoft Visual C# 2017":    struct{}{},
		"Microsoft Visual Basic 2017": struct{}{},
		"PyPy3.9 v7.3.9":              struct{}{},
		"Go 1.16.3":                   struct{}{},
		"Node.js 19.0.0":              struct{}{},
	}

	_, exist := languagesDict[language]

	if !exist {
		return false
	}

	return true
}

func (t *Codeforces) GetLanguages() []string {
	return []string{
		"GNU GCC C11 5.1.0",
		"GNU G++ 14 6.4.0",
		"Python 3.8.10",
		"PascalABC.NET 3.8.3",
		"Java SE JDK 16.0.1",
		"Free Pascal 3.2.2",
		"Borland Delphi 7.0",
		"Microsoft Visual C++ 2017",
		"Microsoft Visual C# 2017",
		"Microsoft Visual Basic 2017",
		"PyPy3.9 v7.3.9",
		"Go 1.16.3",
		"Node.js 19.0.0",
	}
}

func (t *Codeforces) Submitter(wg *sync.WaitGroup, ch chan<- database.Submission) {

	// Создаем новый cookie jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Выполняем первый запрос для получения CSRF-токена
	resp1, err := client.Get("https://codeforces.com/enter?back=%2F")
	defer resp1.Body.Close()

	htmlData, err := ioutil.ReadAll(resp1.Body)
	doc, err := html.Parse(bytes.NewReader(htmlData))

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

	if csrfToken == "" {
		fmt.Println("Не удалось найти CSRF-токен")
	} else {
		fmt.Printf("CSRF-токен: %s\n", csrfToken)
	}

	// Теперь, когда у нас есть CSRF-токен, мы можем использовать его в последующих запросах

	loginURL := "https://Codeforces.com/enter?back=%2F"
	loginData := url.Values{
		"csrf_token":    {csrfToken},
		"action":        {"enter"},
		"ftaa":          {"nxnnepf797p929r93p"},
		"bfaa":          {"939f6b320d3e9e423cd3b4899db9631d"},
		"handleOrEmail": {"gordejka179"},
		"password":      {"XB#8T^m;xj5n~;8"},
		"_tta":          {"661"},
	}

	resp, _ := client.PostForm(loginURL, loginData)

	defer resp.Body.Close()

	fileName := "enter.html"
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}

	err = ioutil.WriteFile(fileName, responseBody, 0644)

	myToCodeforcesDict := map[string]string{
		"GNU GCC C11 5.1.0":           "43",
		"GNU G++ 14 6.4.0":            "50",
		"Python 3.8.10":               "31",
		"PascalABC.NET 3.8.3":         "PP",
		"Java SE JDK 16.0.1":          "JAVA",
		"Free Pascal 3.2.2":           "PAS",
		"Borland Delphi 7.0":          "DPR",
		"Microsoft Visual C++ 2017":   "CXX",
		"Microsoft Visual C# 2017":    "CS",
		"Microsoft Visual Basic 2017": "BAS",
		"PyPy3.9 v7.3.9":              "PYPY",
		"Go 1.16.3":                   "GO",
		"Node.js 19.0.0":              "JS",
	}

	for {
		submissions, err := database.GetSubmitsWithStatus(t.GetName(), 0)
		if err != nil {
			fmt.Println(err)
		}

		// перебираем все решения
		for _, submission := range submissions {
			fileData := url.Values{
				"csrf_token": {csrfToken},
				"ftaa":       {"nxnnepf797p929r93p"},
				"bfaa":       {"939f6b320d3e9e423cd3b4899db9631d"},
				"action":     {"submitSolutionFormSubmitted"},

				"contestId":             {string(strings.Split(submission.TaskID, "/")[0])},
				"submittedProblemIndex": {string(strings.Split(submission.TaskID, "/")[1])},

				"programTypeId": {myToCodeforcesDict[submission.Language]},
				"source":        {string(submission.Code)},
				"tabSize":       {"4"},
				"sourceFile":    {""},
				"_tta":          {"661"},
			}

			submitURL := fmt.Sprintf("https://codeforces.com/problemset/submit/%s", submission.TaskID)

			req, _ := http.NewRequest("POST", submitURL, strings.NewReader(fileData.Encode()))

			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
			req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")

			req.Header.Set("Connection", "keep-alive")
			//req.Header.Set("Content-Length", "какое то число")

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			//req.Header.Set("Cookie", s)

			req.Header.Set("Host", "codeforces.com")

			req.Header.Set("Priority", "u=0, i")

			req.Header.Set("Referer", "https://codeforces.com/problemset/submit/1992/G")
			req.Header.Set("Sec-Fetch-Dest", "document")
			req.Header.Set("Sec-Fetch-Mode", "navigate")

			req.Header.Set("Sec-Fetch-Site", "same-origin")
			req.Header.Set("Sec-Fetch-User", "?1")
			req.Header.Set("TE", "trailers")
			req.Header.Set("Upgrade-Insecure-Requests", "1")

			req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")

			/*
				resp, err = client.Do(req)
				if err != nil {
					fmt.Println(err)
				}
				defer resp.Body.Close()
			*/
			client.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
			resp, err := client.PostForm(submitURL, fileData)

			fileName = "codeforces_response.html"
			responseBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Ошибка при чтении ответа:", err)
				return
			}
			fmt.Println("Response Headers:")
			for k, v := range resp.Header {
				fmt.Printf("%s: %v\n", k, v)
			}

			err = ioutil.WriteFile(fileName, responseBody, 0644)

			// теперь надо передать по каналу, что был изменен статус этой задачи
			submission.Status = 1
			submission.Verdict = "Compiling"
			submission.SubmissionNumber = "0"
			ch <- submission
		}

		time.Sleep(time.Second * 2)
	}

}

func (t *Codeforces) Checker(wg *sync.WaitGroup, ch chan<- database.Submission) {
	// жесткий парсинг таблицы результатов
	defer wg.Done()

	for {
		submissions, err := database.GetSubmitsWithStatus(t.GetName(), 1)

		if err != nil {
			fmt.Println(err)
		}

		submissionsDict := make(map[string]database.Submission)
		submissionsIDs := make([]string, 0)

		for _, submission := range submissions {
			submissionsDict[submission.SubmissionNumber] = submission
			submissionsIDs = append(submissionsIDs, submission.SubmissionNumber)
		}

		pageNum := 0
		for len(submissions) != 0 {
			currentUrl := fmt.Sprintf("https://Codeforces.ru/index.asp?main=status&id_mem=%d&id_res=0&id_t=0&page=%d", 333835, pageNum)

			result, err := http.Get(currentUrl)
			if err != nil {
				fmt.Println("Error:", err)
			}
			defer result.Body.Close()

			utf8Reader, err := decodeWindows1251(result.Body)
			if err != nil {
				log.Fatal(err)
			}

			doc, err := goquery.NewDocumentFromReader(utf8Reader)
			if err != nil {
				log.Fatal(err)
			}

			table := doc.Find("table.main.refresh[align='center']")
			table.Find("tr").Each(func(index int, rowHtml *goquery.Selection) {
				columns := rowHtml.Find("td")
				idStr := columns.Eq(0).Text()

				for _, submissionID := range submissionsIDs {
					if idStr == submissionID {
						// Это строка с нужным id, вы можете выполнить здесь нужные действия 5 6 7 8
						// удаление из словаря и списка
						submission, exists := submissionsDict[idStr]
						if !exists {
							log.Println("Submission with ID not found:", idStr)
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

						verdict := strings.TrimSpace(columns.Eq(5).Text())
						test := strings.TrimSpace(columns.Eq(6).Text())
						executionTime := strings.TrimSpace(columns.Eq(7).Text())
						memoryUsed := strings.TrimSpace(columns.Eq(8).Text())

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
			})

			pageNum++
		}

		time.Sleep(time.Second * 2)
	}
}

func (t *Codeforces) GetProblem(taskID string) (database.Task, error) {
	taskURL := fmt.Sprintf("https://codeforces.com/problemset/problem/%s/?locale=ru", taskID)

	resp, err := http.Get(taskURL)
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
			taskName = s.Text()[2:]
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
