package database

import (
	"context"

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
		"INSERT INTO forumtable (forumtheme, forumtask, forumuser) VALUES ($1, $2, $3) RETURNING id",
		msg.ForumTheme,
		msg.ForumTask,
		msg.ForumUser,
	).Scan(&contestID)
	if err != nil {
		return 0, err
	}

	return contestID, nil
}
