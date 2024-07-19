package database

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
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

func GetContestInfoById(id int) (Contest, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return Contest{}, err
	}
	defer conn.Close(context.Background())

	var contest Contest
	err = conn.QueryRow(context.Background(), "SELECT * FROM contests WHERE id = $1", id).Scan(
		&contest.ID,
		&contest.Title,
		&contest.Access,
		&contest.Participants,
		&contest.Results,
		&contest.Tasks,
		&contest.GroupOwner,
		&contest.StartTime,
		&contest.Duration,
	)
	if err != nil {
		return Contest{}, err
	}

	return contest, nil
}

func CreateContest(title string, access map[string]interface{}, groupOwner int, startTime time.Time, duration time.Duration) (int, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return 0, err
	}
	defer conn.Close(context.Background())

	var contestID int
	err = conn.QueryRow(
		context.Background(),
		"INSERT INTO contests (title, access, group_owner, start_time, duration) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		title,
		access,
		groupOwner,
		startTime,
		duration,
	).Scan(&contestID)
	if err != nil {
		return 0, err
	}

	return contestID, nil
}
