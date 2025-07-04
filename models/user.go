package models

import (
	"errors"
	"go-multiplayer-quiz-project/backend/database"
	"go-multiplayer-quiz-project/backend/utils"

	"github.com/lib/pq"
)

type User struct {
	UserId   int64
	Username string
	Password string
}

func (user *User) SaveUserToDB() error {

	userQuery := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING userid"

	err := database.DB.QueryRow(userQuery, user.Username, user.Password).Scan(&user.UserId)
	if err != nil {
		// Check if the error is the unique voilation error
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return errors.New("username already exists")
		}
		return err
	}

	playerQuery := "INSERT INTO players (playerid, username) VALUES ($1, $2)"
	_, err = database.DB.Exec(playerQuery, user.UserId, user.Username)
	if err != nil {
		return err
	}

	return nil
}

func (user *User) ValidateLogin() (err error) {

	query := " SELECT * FROM users WHERE username = $1 "

	row := database.DB.QueryRow(query, user.Username)

	var hashedPassword string

	err = row.Scan(&user.UserId, &user.Username, &hashedPassword)
	if err != nil {
		return errors.New("invalid Username/Password")
	}

	match := utils.CheckPassword(hashedPassword, user.Password)

	if !match {
		return errors.New("invalid Username/Password")
	}

	return err
}
