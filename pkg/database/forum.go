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

func GetAllUserThemes(user string) ([]string, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	mem := make(map[string]bool)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT forumtheme FROM forumtable WHERE forumuser = $1", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []string
	for rows.Next() {
		var res1 string
		if err := rows.Scan(&res1); err != nil {
			return nil, err
		}
		if !mem[res1] {
			res = append(res, res1)
			mem[res1] = true
		}
	}

	return res, nil
}

type messatheme struct {
	theme string
	msgs  []messa
}

func GetUserForum(user string) ([]messatheme, []string, []string, []string, []string, error) { //1-theme,2-task,3-user,4-time
	thems, err := GetAllUserThemes(user)
	res := make([]messatheme, 0)
	theme, task, usern, ti := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	for _, them := range thems {
		msgs, _, err := GetTasksFromTheme(them)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		for _, m := range msgs {
			theme = append(theme, m.theme)
			task = append(task, m.task)
			usern = append(usern, m.sender)
			ti = append(ti, m.time)
		}
		res1 := messatheme{
			theme: them,
			msgs:  msgs,
		}
		res = append(res, res1)
	}
	return res, theme, task, usern, ti, nil
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
