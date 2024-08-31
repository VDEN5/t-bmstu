package database

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/jackc/pgx/v4"
)

func CreateSolution(sol Ranktable) (int, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return 0, err
	}
	defer conn.Close(context.Background())

	var solID int
	err = conn.QueryRow(
		context.Background(),
		"INSERT INTO ranktable (rankuser, testsystem, problemid,points,time) VALUES ($1, $2, $3, $4,$5) RETURNING id",
		sol.RankUser,
		sol.TestSystem,
		sol.ProblemId,
		sol.Points,
		sol.Time,
	).Scan(&solID)
	if err != nil {
		return 0, err
	}

	return solID, nil
}

func GetUserSols(rankuser string) ([]Ranktable, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)

	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	query := `
		SELECT id, rankuser, testsystem, problemid, points
		FROM ranktable
		WHERE rankuser = $1
	`

	rows, err := conn.Query(context.Background(), query, rankuser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verdicts []Ranktable
	for rows.Next() {
		var verdict Ranktable
		err := rows.Scan(&verdict.ID, &verdict.RankUser, &verdict.TestSystem, &verdict.ProblemId, &verdict.Points)
		if err != nil {
			log.Printf("Failed to fetch submission verdict: %v", err)
		}

		verdicts = append(verdicts, verdict)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return verdicts, nil
}

func ExistSols(login string, testsystem string, id string) (bool, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return false, err
	}
	defer conn.Close(context.Background())

	var count int

	query := `
  SELECT COUNT(*) FROM ranktable WHERE rankuser = $1 AND testsystem = $2 AND problemid = $3;
 `

	err = conn.QueryRow(context.Background(), query, login, testsystem, id).
		Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetRankInfo() (map[string]int, error) { //1-theme,2-task,3-user,4-time
	conn, err := pgx.Connect(context.Background(), DbURL)

	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	query := `
		SELECT rankuser, points
		FROM ranktable
		ORDER BY id ASC
	`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int)
	for rows.Next() {
		var user string
		var points int
		err := rows.Scan(&user, &points)
		if err != nil {
			log.Printf("Failed to fetch submission verdict: %v", err)
		}

		res[user] += points
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAllSols(user string) ([]int, error) {
	conn, err := pgx.Connect(context.Background(), DbURL)
	mem := make(map[int]bool)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT problemid FROM ranktable WHERE rankuser = $1", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []int
	for rows.Next() {
		var res1 int
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

func GetSol(user string, testsystem string, problemidstr string) (Ranktable, error) {
	problemid, err := strconv.Atoi(problemidstr)
	if err != nil {
		return Ranktable{}, fmt.Errorf("this is not number: %v", err)
	}

	conn, err := pgx.Connect(context.Background(), DbURL)
	if err != nil {
		return Ranktable{}, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	var let Ranktable
	err = conn.QueryRow(context.Background(), "SELECT id, rankuser, testsystem, problemid, time, points FROM ranktable WHERE rankuser = $1 AND testsystem = $2 AND problemid = $3", user, testsystem, problemid).
		Scan(&let.ID, &let.RankUser, &let.TestSystem, &let.ProblemId, &let.Time, &let.Points)
	if err != nil {
		return Ranktable{}, fmt.Errorf("query execution failed: %v", err)
	}

	return let, nil
}
