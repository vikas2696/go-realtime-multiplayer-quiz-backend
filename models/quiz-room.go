package models

import (
	"encoding/json"
	"errors"
	"go-multiplayer-quiz-project/backend/database"
)

type QuizRoom struct {
	QuizRoomId     int64
	Players        []Player
	TimerTime      int
	QuizTopic      string
	IsRunnning     bool
	ScoreSheet     map[int64]int
	PlayersAnswers map[int64]string
}

func GetQuizRoomsFromDB() ([]QuizRoom, error) {

	var q []QuizRoom

	query := "SELECT * FROM quizrooms"

	rows, err := database.DB.Query(query)

	if err != nil {
		return q, err
	}

	defer rows.Close()

	var quizRoom QuizRoom
	var playersData string
	var scoresheetData string
	var playersAnswersData string
	var playersList []Player
	for rows.Next() {

		quizRoom = QuizRoom{}
		playersData = ""
		scoresheetData = ""
		playersAnswersData = ""
		playersList = nil

		err = rows.Scan(&quizRoom.QuizRoomId, &playersData, &quizRoom.TimerTime, &quizRoom.QuizTopic, &quizRoom.IsRunnning, &scoresheetData, &playersAnswersData)

		if err != nil {
			return q, err
		}

		err = json.Unmarshal([]byte(playersData), &playersList)
		if err != nil {
			return q, err
		}

		err = json.Unmarshal([]byte(playersAnswersData), &quizRoom.PlayersAnswers)
		if err != nil {
			return q, err
		}

		err = json.Unmarshal([]byte(scoresheetData), &quizRoom.ScoreSheet)
		if err != nil {
			return q, err
		}

		quizRoom.Players = playersList

		q = append(q, quizRoom)
	}

	return q, err
}

func (quizRoom QuizRoom) SaveQuizRoomToDB() (quizRoomId int64, err error) {

	query := `
		INSERT INTO quizrooms (players, timertime, quiztopic, isrunning, scoresheet, playersanswers) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING quizroomid
	`

	playersJson, err := json.Marshal(quizRoom.Players)
	if err != nil {
		return 0, err
	}

	hostPlayerId := quizRoom.Players[0].PlayerId

	quizRoom.ScoreSheet = make(map[int64]int)
	quizRoom.ScoreSheet[hostPlayerId] = 0

	scoreSheetJson, err := json.Marshal(quizRoom.ScoreSheet)
	if err != nil {
		return 0, err
	}

	playersAnswersJson, err := json.Marshal(quizRoom.PlayersAnswers)
	if err != nil {
		return 0, err
	}

	err = database.DB.QueryRow(
		query,
		playersJson,
		quizRoom.TimerTime,
		quizRoom.QuizTopic,
		quizRoom.IsRunnning,
		scoreSheetJson,
		playersAnswersJson,
	).Scan(&quizRoomId)

	if err != nil {
		return 0, err
	}

	return quizRoomId, err
}

func (q *QuizRoom) GetQuizRoomFromId(quizId int) error {

	query := `	SELECT * FROM quizrooms WHERE quizroomid = $1  `

	rows := database.DB.QueryRow(query, quizId)

	var playersData string
	var scoresheetData string
	var playersanswersData string

	err := rows.Scan(&q.QuizRoomId, &playersData, &q.TimerTime, &q.QuizTopic, &q.IsRunnning, &scoresheetData, &playersanswersData)
	if err != nil {
		return errors.New("unable to scan rows")
	}

	err = json.Unmarshal([]byte(playersData), &q.Players)
	if err != nil {
		return errors.New("unable to UnMarshal players")
	}

	err = json.Unmarshal([]byte(scoresheetData), &q.ScoreSheet)
	if err != nil {
		return errors.New("unable to unmarshal scoresheet")
	}

	err = json.Unmarshal([]byte(playersanswersData), &q.PlayersAnswers)
	if err != nil {
		return errors.New("unable to unmarshal scoresheet")
	}

	return err
}

func UpdateRoomStatus(quizRoomId int64) error {

	query := "	UPDATE quizrooms SET isrunning = $1 WHERE quizroomid = $2"

	result, err := database.DB.Exec(query, true, quizRoomId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("invalid Quiz Room")
	}

	return err

}

func AddPlayerToScoreSheet(quizRoomId int, playerId int64, scoreSheet map[int64]int) error {

	query := "	UPDATE quizrooms SET scoresheet = $1 WHERE quizroomid = $2  "

	scoreSheet[playerId] = 0

	scoreSheetData, err := json.Marshal(scoreSheet)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, scoreSheetData, quizRoomId)
	if err != nil {
		return err
	}

	return err

}

func RemovePlayerFromScoreSheet(quizRoomId int, playerId int64, scoreSheet map[int64]int) error {

	query := "	UPDATE quizrooms SET scoresheet = $1 WHERE quizroomid = $2  "

	delete(scoreSheet, playerId)

	scoreSheetData, err := json.Marshal(scoreSheet)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, scoreSheetData, quizRoomId)
	if err != nil {
		return err
	}

	return err

}

func AddPlayerToPlayersAnswers(quizRoomId int, playerId int64, playersAnswers map[int64]string) error {

	query := "	UPDATE quizrooms SET playersanswers = $1 WHERE quizroomid = $2  "

	playersAnswers[playerId] = ""

	playersAnswersData, err := json.Marshal(playersAnswers)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, playersAnswersData, quizRoomId)
	if err != nil {
		return err
	}

	return err

}

func RemovePlayerToPlayersAnswers(quizRoomId int, playerId int64, playersAnswers map[int64]string) error {

	query := "	UPDATE quizrooms SET playersanswers = $1 WHERE quizroomid = $2  "

	delete(playersAnswers, playerId)

	playersAnswersData, err := json.Marshal(playersAnswers)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, playersAnswersData, quizRoomId)
	if err != nil {
		return err
	}

	return err

}

func SaveAnswersToDB(playersAnswers map[int64]string, quizRoomId int) error {

	query := "UPDATE quizrooms SET playersanswers = $1  WHERE quizroomid = $2"

	playersAnswersData, err := json.Marshal(playersAnswers)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, playersAnswersData, quizRoomId)
	if err != nil {
		return err
	}

	return err

}

func UpdateScoreSheetinDB(quizRoomId int64, scoreSheet map[int64]int) error {
	query := "UPDATE quizrooms SET scoresheet = $1  WHERE quizroomid = $2"

	scoreSheetData, err := json.Marshal(scoreSheet)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(query, scoreSheetData, quizRoomId)
	if err != nil {
		return err
	}

	return err
}

func DeleteQuizRoomFromDB(quizId int64) error {

	query := "DELETE FROM quizrooms WHERE quizroomid = $1"

	_, err := database.DB.Exec(query, quizId)
	if err != nil {
		return errors.New("invalid deletion")
	}

	return err
}

func IsHost(quizId int, playerId int64) bool {

	var quizRoom QuizRoom
	quizRoom.GetQuizRoomFromId(quizId)

	return playerId == quizRoom.Players[0].PlayerId
}
