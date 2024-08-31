package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	//"time"
)

func CreateLet(let Letuchka) (int, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return 0, err
	}
	defer conn.Close(context.Background())

	var contestID int
	err = conn.QueryRow(
		context.Background(),
		"INSERT INTO lettable (tasks, soltasks, letuser,starttime, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		let.Tasks,
		let.Soltasks,
		let.Letuser,
		let.Starttime,
		let.Status, //1-идет сейчас, 2-ждет утверждения результатов, 3-результаты утверждены
	).Scan(&contestID)
	if err != nil {
		return 0, err
	}

	return contestID, nil
}

func ExistLet(login string, stat int) (bool, error) { //есть ли летучка
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return false, err
	}
	defer conn.Close(context.Background())

	var count int

	query := `
  SELECT COUNT(*) FROM lettable WHERE letuser = $1 AND status = $2;
 `

	err = conn.QueryRow(context.Background(), query, login, stat).
		Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetLet(user string, status int) (Letuchka, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return Letuchka{}, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	var let Letuchka
	err = conn.QueryRow(context.Background(), "SELECT id, tasks, soltasks, letuser, starttime, status FROM lettable WHERE letuser = $1 AND status = $2", user, status).Scan(&let.ID, &let.Tasks, &let.Soltasks, &let.Letuser, &let.Starttime, &let.Status)
	if err != nil {
		return Letuchka{}, fmt.Errorf("query execution failed: %v", err)
	}

	return let, nil
}

func Updstattofin(let Letuchka) error {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := `
        UPDATE lettable
        SET
            status = $1
        WHERE id = $2
    `

	_, err = conn.Exec(
		context.Background(),
		query,
		let.Status,
		let.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func Updsols(let Letuchka) error {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := `
        UPDATE lettable
        SET
            soltasks = $1
        WHERE id = $2
    `

	_, err = conn.Exec(
		context.Background(),
		query,
		let.Soltasks,
		let.ID,
	)
	if err != nil {
		return err
	}

	return nil
}
