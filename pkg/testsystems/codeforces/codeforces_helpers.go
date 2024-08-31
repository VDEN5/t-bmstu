package codeforces

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

func saveToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func extractValueSuf(part, prefix string) string {
	index := strings.Index(part, prefix)
	if index >= 0 {
		value := strings.TrimSpace(part[:index])
		return value
	}
	return ""
}

func extractValuePref(part, prefix string) string {
	index := strings.Index(part, prefix)
	if index >= 0 {
		value := strings.TrimSpace(part[index+len(prefix):])
		return strings.Replace(value, " сек.", "", -1)
	}
	return ""
}

func endChecking(verdict string) bool {
	if verdict == "Compilation error" || verdict == "Wrong answer" || verdict == "Accepted" ||
		verdict == "Time limit exceeded" || verdict == "Memory limit exceeded" || verdict == "Runtime error (non-zero exit code)" ||
		verdict == "RUNTIME_ERROR" || verdict == "WRONG_ANSWER" || verdict == "OK" || verdict == "COMPILATION_ERROR" {
		return true
	}
	return false
}

func removeLeadingZeros(s string) string {
	trimmed := strings.TrimLeft(s, "0")
	if trimmed == "" {
		return "0"
	}
	return trimmed
}

type Task struct {
	ID      string
	Name    string
	Points  string
	Yourpoi int
}

func parseTableToJSON(table *goquery.Selection, doc *goquery.Document) string {
	tests := []map[string]string{}
	inputData := ""
	outputData := ""

	doc.Find("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests div.sample-test div.input").NextUntil("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests div.sample-test div.input").Each(func(i int, s *goquery.Selection) {
		outputData = s.Text()
		outputData = strings.TrimPrefix(outputData, "Выходные данные")
	})

	table.Find(".test-example-line").Each(func(i int, s *goquery.Selection) {
		inputData += s.Text() + "\n"
	})
	var test map[string]string
	if strings.TrimSpace(inputData) == "" || strings.TrimSpace(outputData) == "" {
		test = map[string]string{
			"input":  "Для этой задачи нет тестов",
			"output": "-",
		}

	} else {
		test = map[string]string{
			"input":  strings.TrimSpace(inputData),
			"output": strings.TrimSpace(outputData),
		}

	}
	tests = append(tests, test)
	jsonTests, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return string(jsonTests)
}

func GetTaskList(from, count int) ([]Task, error) {
	req, err := http.NewRequest("GET", "https://codeforces.com/problemset/?locale=ru", nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	result, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer result.Body.Close()

	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	doc, err := goquery.NewDocumentFromReader(result.Body)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	taskCount := 0

	doc.Find("table.problems tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		contestIDStr := s.Find("td").Eq(0).Text()
		nameStr := s.Find("td").Eq(1).Text()

		ID := strings.TrimSpace(contestIDStr)
		name := strings.TrimSpace(nameStr)

		firstLetterIndex := strings.IndexFunc(ID, unicode.IsLetter)
		if firstLetterIndex != -1 {
			ID = ID[:firstLetterIndex] + "/" + ID[firstLetterIndex:]

		}

		// Находим первый перевод строки в problemID
		newlineIndex := strings.IndexByte(name, '\n')
		if newlineIndex != -1 {
			name = name[:newlineIndex]
		}

		if ID != "" && name != "" {
			idBytes := []byte("codeforces" + ID)
			tasks = append(tasks, Task{
				ID:   base64.StdEncoding.EncodeToString(idBytes),
				Name: name,
			})
			taskCount++
		}

		if taskCount >= count {
			return
		}
	})
	return tasks[from-1 : from+count-1], nil
}
