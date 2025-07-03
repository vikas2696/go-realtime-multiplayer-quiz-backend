package models

import (
	"encoding/json"
	"errors"
	"go-multiplayer-quiz-project/backend/database"
)

type Player struct {
	PlayerId int64
	Username string
}

func GetJoinedPlayersList(quizId int) ([]Player, error) {

	var players []Player
	var dataString string

	query := "SELECT players FROM quizrooms WHERE quizroomid = $1"
	rows, err := database.DB.Query(query, quizId)

	if err != nil {
		return players, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&dataString)
		if err != nil {
			return players, err
		}
	}

	err = json.Unmarshal([]byte(dataString), &players)

	if err != nil {
		return players, err
	}

	return players, err

}

func (player Player) AddPlayerToQuiz(quizId int) error {

	var quizRoom QuizRoom
	err := quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		return errors.New("quizroom not found")
	}

	if quizRoom.IsRunnning {
		return errors.New("quiz is already going on")
	}

	query := " UPDATE quizrooms SET  players = $1 WHERE quizroomid = $2 "

	players, err := GetJoinedPlayersList(quizId)
	if err != nil {
		return err
	}

	for index := range players {

		if players[index].PlayerId == player.PlayerId {
			return errors.New("player already joined")
		}

	}

	players = append(players, player)

	playersJson, err := json.Marshal(players)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, playersJson, quizId)

	if err != nil {
		return err
	}

	err = AddPlayerToScoreSheet(quizId, player.PlayerId, quizRoom.ScoreSheet)
	if err != nil {
		return errors.New("unable to add player to scoresheet")
	}

	err = AddPlayerToPlayersAnswers(quizId, player.PlayerId, quizRoom.PlayersAnswers)
	if err != nil {
		return errors.New("unable to add player to scoresheet")
	}

	return err
}

func (player Player) DeletePlayerFromQuiz(quizId int) error {

	var quizRoom QuizRoom
	err := quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		return err
	}

	players, err := GetJoinedPlayersList(quizId)
	if err != nil {
		return err
	}

	indexToRemove := -1
	for index := range players {

		if players[index].PlayerId == player.PlayerId {
			indexToRemove = index
			break
		}

	}

	if indexToRemove != -1 {
		players = append(players[:indexToRemove], players[indexToRemove+1:]...)
	} else {
		return errors.New("invalid Request")
	}

	playersJson, err := json.Marshal(players)
	if err != nil {
		return err
	}
	query := " UPDATE quizrooms SET players = $1 WHERE quizroomid = $2 "

	_, err = database.DB.Exec(query, playersJson, quizId)

	if err != nil {
		return err
	}

	err = RemovePlayerFromScoreSheet(quizId, player.PlayerId, quizRoom.ScoreSheet)
	if err != nil {
		return errors.New("unable to remove player to scoresheet")
	}

	return err
}

func GetPlayerFromId(pId int) (p Player, err error) {

	query := "SELECT * FROM players WHERE playerid = $1"
	rows, err := database.DB.Query(query, pId)

	if err != nil {
		return p, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&p.PlayerId, &p.Username)
		if err != nil {
			return p, err
		}
	}

	return p, err
}
