package database

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB
var err error

func init() {
	_ = godotenv.Load(".env") // only used locally
}

func InitDB() {

	dbUrl := os.Getenv("DATABASE_URL")
	DB, err = sql.Open("postgres", dbUrl)

	if err != nil {
		panic("Cannot connect to the database: " + err.Error())
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	createTables()
}

func createTables() {

	quizRoomQuery := `
		CREATE TABLE IF NOT EXISTS quizrooms (
			quizroomid SERIAL PRIMARY KEY,
			players TEXT,
			timertime INTEGER NOT NULL,
			quiztopic TEXT NOT NULL,
			isrunning BOOLEAN NOT NULL DEFAULT false,
			scoresheet TEXT NOT NULL,
			playersanswers TEXT NOT NULL DEFAULT ''
		)`

	_, err := DB.Exec(quizRoomQuery)

	if err != nil {
		panic("Unable to initialize quiz room table" + err.Error())
	}

	playerQuery := `
		CREATE TABLE IF NOT EXISTS players (
		playerid INTEGER PRIMARY KEY,
		username TEXT NOT NULL
		)`

	_, err = DB.Exec(playerQuery)
	if err != nil {
		panic("Unable to initialize player table" + err.Error())
	}

	userQuery := `
		CREATE TABLE IF NOT EXISTS users (
		userid SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
		)`

	_, err = DB.Exec(userQuery)
	if err != nil {
		panic("Unable to initialize user table" + err.Error())
	}

}
