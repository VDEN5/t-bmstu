package codeforces

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func decodeWindows1251(reader io.Reader) (io.Reader, error) {
	decoder := charmap.Windows1251.NewDecoder()
	return transform.NewReader(reader, decoder), nil
}

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

func parseTableToJSON(table *goquery.Selection, doc *goquery.Document) string {
	tests := []map[string]string{}
	inputData := ""
	outputData := ""
	/* у codeforces какая то неадекватная структура.
	В том числе не всегда даны входные и выходные данные,
	*/
	doc.Find("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests div.sample-test div.input").NextUntil("html body div#body div div#pageContent.content-with-sidebar div.problemindexholder div.ttypography div.problem-statement div.sample-tests div.sample-test div.input").Each(func(i int, s *goquery.Selection) {
		outputData = s.Text()
		outputData = strings.TrimPrefix(outputData, "Выходные данные")
	})

	table.Find(".test-example-line").Each(func(i int, s *goquery.Selection) {
		inputData += s.Text() + "\n"
	})
	test := map[string]string{
		"input":  strings.TrimSpace(inputData),
		"output": strings.TrimSpace(outputData),
	}
	tests = append(tests, test)
	jsonTests, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return string(jsonTests)
}

// Submit - Функция для выполнения второго запроса
func Submit(client *http.Client, submitURL string, fileData url.Values, taskId string) (string, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Добавление остальных полей в multipart форму
	for key, val := range fileData {
		_ = writer.WriteField(key, val[0])
	}

	// Закрытие multipart формы
	writer.Close()

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", submitURL, &buffer)
	if err != nil {
		return "0", err
	}

	// Установка заголовка Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправка запроса
	resp, err := client.Do(req)

	if err != nil {
		return "0", err
	}
	defer resp.Body.Close()

	/*
		forIdUrl := fmt.Sprintf("https://codeforces.ru/index.asp?main=status&id_mem=%d&id_res=0&id_t=%s&page=0", 333835, taskId)
		result, err := http.Get(forIdUrl)
		if err != nil {
			fmt.Println("Error:", err)
			return "1", err
		}
		defer result.Body.Close()

		utf8Reader, err := decodeWindows1251(result.Body)
		if err != nil {
			log.Fatal(err)
			return "1", err
		}

		doc, err := goquery.NewDocumentFromReader(utf8Reader)
		if err != nil {
			log.Fatal(err)
			return "1", err
		}

		table := doc.Find("table.main.refresh[align='center']")
		if table.Length() > 0 {
			// Найти первую строку таблицы, которая не является заголовком
			rows := table.Find("tr")
			for i := 1; i < rows.Length(); i++ { // Начинаем с 1, чтобы пропустить заголовок
				row := rows.Eq(i)
				columns := row.Find("td")
				if columns.Length() > 0 {
					id := columns.Eq(0).Text()
					fmt.Sprintf(id)
					return id, nil
				}
			}
		} else {
			fmt.Println("Table not found")
		}
	*/

	return "1", nil
}

func endChecking(verdict string) bool {
	if verdict == "Compilation error" || verdict == "Wrong answer" || verdict == "Accepted" ||
		verdict == "Time limit exceeded" || verdict == "Memory limit exceeded" || verdict == "Runtime error (non-zero exit code)" ||
		verdict == "Runtime error" {
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
	ID   string
	Name string
}

func GetTaskList(count int) ([]Task, error) {
	result, err := http.Get("https://codeforces.com/problemset/?locale=ru")
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
	return tasks, nil
}
