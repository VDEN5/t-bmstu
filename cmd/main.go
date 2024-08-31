package main

import (
	"log"
	"os"

	t_bmstu "github.com/VDEN5/t-bmstu"
	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/VDEN5/t-bmstu/pkg/handler"
	"github.com/VDEN5/t-bmstu/pkg/testsystems"
)

type Config struct {
	appPort       string
	dbUsername    string
	dbPassword    string
	dbName        string
	SessionSecret string
	dbHost        string
}

func main() {
	conf := InitConfig()

	err := database.CreateTables(conf.dbUsername, conf.dbPassword, conf.dbHost, conf.dbName)

	if err != nil {
		log.Fatalf("Error occured while creating tables: %s", err.Error())
	}

	// normal code
	handlers := new(handler.Handler)

	srv := new(t_bmstu.Server)

	// запуск горутин проверки задач
	go testsystems.InitGorutines()

	if err := srv.Run(conf.appPort, handlers.InitRoutes()); err != nil {
		log.Fatalf("Error occured while running http server: %s", err.Error())
	}

}

func InitConfig() Config {
	result := Config{
		appPort:       os.Getenv("APP_PORT"),
		dbUsername:    os.Getenv("DB_USER"),
		dbPassword:    os.Getenv("DB_PASSWORD"),
		dbName:        os.Getenv("DB_NAME"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		dbHost:        os.Getenv("DB_HOST"),
	}

	// default value
	if result.dbHost == "" {
		return Config{
			appPort:       "8080",
			dbUsername:    "postgres",
			dbPassword:    "qwerty", //тут укажи свой пароль
			dbName:        "postgres",
			SessionSecret: "session",
			dbHost:        "localhost",
		}
	}

	return result
}
