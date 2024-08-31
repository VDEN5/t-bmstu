package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	//"time"
)

func CreateMessage(msg Message) (int, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return 0, err
	}
	defer conn.Close(context.Background())

	var contestID int
	err = conn.QueryRow(
		context.Background(),
		"INSERT INTO forumtable (forumtheme, forumtask, forumuser,msgtime) VALUES ($1, $2, $3, $4) RETURNING id",
		msg.ForumTheme,
		msg.ForumTask,
		msg.ForumUser,
		msg.MSGtime,
	).Scan(&contestID)
	if err != nil {
		return 0, err
	}

	return contestID, nil
}

func GetTheme(id int) (string, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	var msg Message
	err = conn.QueryRow(context.Background(), "SELECT forumtheme FROM forumtable WHERE id = $1", id).Scan(&msg.ForumTheme)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %v", err)
	}

	return string(msg.ForumTheme), nil
}

func GetTask(id int) (string, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	var msg Message
	err = conn.QueryRow(context.Background(), "SELECT forumtask FROM forumtable WHERE id = $1", id).Scan(&msg.ForumTask)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %v", err)
	}

	return msg.ForumTask, nil
}

type messa struct {
	theme  string
	task   string
	sender string
	time   string
}

func GetTasksFromTheme(theme string) ([]messa, []string, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT forumtheme,forumuser,forumtask,msgtime FROM forumtable")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var res []messa
	var r []string
	for rows.Next() {
		var res1 messa
		var r1 string
		theme1 := ""
		if err := rows.Scan(&theme1, &res1.sender, &res1.task, &res1.time); err != nil {
			return nil, nil, err
		}
		if (theme1 == theme) && (res1.task != "") {
			res1.theme = theme1
			res = append(res, res1)
			r = append(r, r1)
		}
	}

	return res, r, nil
}

type mes struct {
	task   string `json:"forumtask"`
	sender string `json:"fsender"`
	time   string `json:"ftime"`
}

func GetTasksFromTheme1(theme string) ([]string, []string, []string, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT forumtheme,forumtask,forumuser,msgtime FROM forumtable")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var ta, se, ti []string
	for rows.Next() {
		var ta1, se1, ti1 string
		theme1 := ""
		if err := rows.Scan(&theme1, &ta1, &se1, &ti1); err != nil {
			return nil, nil, nil, err
		}
		if (theme1 == theme) && (ta1 != "") {
			ta = append(ta, ta1)
			se = append(se, se1)
			ti = append(ti, ti1)
		}
	}

	return ta, se, ti, nil
}

func GetAllUserThemes(user string) (map[string]Message, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	//mem := make(map[string]bool)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT forumtheme,forumtask,forumuser,msgtime FROM forumtable WHERE forumuser = $1", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]Message)
	for rows.Next() {
		var th, ta, us, ti string
		if err := rows.Scan(&th, &ta, &us, &ti); err != nil {
			return nil, err
		}
		res[th] = Message{
			ForumTheme: th,
			ForumTask:  ta,
			ForumUser:  us,
			MSGtime:    ti,
		}
	}

	return res, nil
}

type messatheme struct {
	theme string
	msgs  []messa
}

func GetInfoForForumProfilePage(id string) (string, string, string) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		fmt.Println(err)
		return "", "", ""
	}
	defer conn.Close(context.Background())
	gituser, fi, la := "", "", ""
	err = conn.QueryRow(context.Background(), "SELECT name3, first_name, last_name FROM users WHERE username = $1", id).Scan(&gituser, &fi, &la)

	if err != nil {
		fmt.Println(err)
		return "", "", ""
	}
	return gituser, fi, la
}
