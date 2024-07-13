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

func MessageExist(msgID string) (bool, Message, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	// Выполнение запроса на проверку наличия сообщения в таблице
	var msg Message
	err = conn.QueryRow(context.Background(), "SELECT * FROM forumtable WHERE id = $1", msgID).Scan(
		&msg.ID,
		&msg.ForumTheme,
		&msg.ForumTask,
		&msg.ForumUser,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, Message{}, nil
		}
		return false, Message{}, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}

	return true, msg, nil
}
