package routeshandlers

import (
	"encoding/json"
	"fmt"
	"go-multiplayer-quiz-project/backend/models"
	"go-multiplayer-quiz-project/backend/utils"
	"log"
	"time"

	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	livePlayers          = make(map[int][]*websocket.Conn)
	live_mu              sync.RWMutex
	live_broadcast_mu    sync.Mutex
	live_broadcastChans  = make(map[int]chan models.LiveMessage)
	questionsPerRoom     = make(map[int][]models.Question)
	current_ques_indices = make(map[int]int)
	scorecardPerRoom     = make(map[int]map[int]*models.PlayerScore)
	tickerChanPerRoom    = make(map[int]*time.Ticker)
	timeLeftPerRoom      = make(map[int]int)
)

func webSocketLive(context *gin.Context) {
	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "unable to load room"})
		return
	}

	var quizRoom models.QuizRoom
	err = quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load room"})
		return
	}

	token := context.Query("token")
	err = utils.ValidateToken(token, context)
	if err != nil {
		context.Writer.WriteHeader(http.StatusUnauthorized)
		context.Writer.Write([]byte("Invalid token"))
		return
	}

	playerId, found := context.Get("userId")
	if !found {
		context.Writer.WriteHeader(http.StatusUnauthorized)
		context.Writer.Write([]byte("player id not found"))
		return
	}

	playerIdInt64, ok := playerId.(int64)
	if !ok {
		log.Println("playerId is not of type int64:", playerId)
		context.Writer.WriteHeader(http.StatusInternalServerError)
		context.Writer.Write([]byte("Invalid player ID type"))
		return
	}

	currentPlayer, err := models.GetPlayerFromId(int(playerIdInt64))
	if err != nil {
		log.Println(err.Error())
		context.Writer.WriteHeader(http.StatusInternalServerError)
		context.Writer.Write([]byte("cannot fetch player info"))
		return
	}

	conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upgrade the connection"})
	}

	live_mu.Lock()
	_, exists := live_broadcastChans[quizId]
	if !exists { //first one to join
		live_broadcastChans[quizId] = make(chan models.LiveMessage)
	}
	if scorecardPerRoom[quizId] == nil {
		scorecardPerRoom[quizId] = make(map[int]*models.PlayerScore)
		timeLeftPerRoom[quizId] = quizRoom.TimerTime
		current_ques_indices[quizId] = -1
		for _, player := range quizRoom.Players {
			scorecardPerRoom[quizId][int(player.PlayerId)] = &models.PlayerScore{
				Username:      player.Username,
				CurrentAnswer: "",
				CurrentScore:  0,
			}
		}
	}
	live_mu.Unlock()

	go livebraodcastAll(quizId)
	go readliveMessages(conn, quizId, quizRoom)

	live_mu.Lock()
	livePlayers[quizId] = append(livePlayers[quizId], conn)
	live_mu.Unlock()
	live_broadcastChans[quizId] <- models.LiveMessage{Type: "join", Msg: currentPlayer, Conn: conn}
}

func livebraodcastAll(quizId int) {

	for { //broadcast code
		liveMessage := <-live_broadcastChans[quizId] //channel to trigger when message receives

		live_mu.RLock()
		conns := append([]*websocket.Conn{}, livePlayers[quizId]...)
		live_mu.RUnlock()

		live_broadcast_mu.Lock()
		for _, connection := range conns { // per room
			err := connection.WriteJSON(liveMessage)
			if err != nil {
				fmt.Println("websocket write error", err)
				connection.Close()
			}
		}
		live_broadcast_mu.Unlock()
	}
}

func readliveMessages(conn *websocket.Conn, quizId int, quizRoom models.QuizRoom) { // to read messages from the frontend

	var clientMsg models.LiveMessage
	for { // per connection
		_, clientMsgJSON, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Connection is closed", err)
			//conn.Close()
			return
		}

		err = json.Unmarshal(clientMsgJSON, &clientMsg)
		if err != nil {
			fmt.Println("Wrong message format: ", err)
		}

		clientMsg.Conn = conn

		switch clientMsg.Type {
		case "questions":
			questionsBytes, err := json.Marshal(clientMsg.Msg)
			if err != nil {
				fmt.Println("Wrong message format: ", err)
			}
			var questions []models.Question
			err = json.Unmarshal(questionsBytes, &questions)
			if err != nil {
				fmt.Println("Wrong message format: ", err)
			}
			if current_ques_indices[quizId] == -1 {
				live_mu.Lock()
				questionsPerRoom[quizId] = questions
				current_ques_indices[quizId] = 0
				live_mu.Unlock()
			}

			live_mu.RLock()
			next_question_available := current_ques_indices[quizId]+1 < len(questionsPerRoom[quizId])
			last_question := current_ques_indices[quizId]+1 == len(questionsPerRoom[quizId])
			question_to_send := questionsPerRoom[quizId][current_ques_indices[quizId]]
			question_to_send.QuestionId = current_ques_indices[quizId] + 1
			data_to_send := gin.H{
				"Question": question_to_send,
				"Timer":    timeLeftPerRoom[quizId],
			}
			live_mu.RUnlock()

			if next_question_available {
				go startTicking(quizId)
				live_broadcastChans[quizId] <- models.LiveMessage{
					Type: "question",
					Msg:  data_to_send,
					Conn: conn}
			} else if last_question {
				go startTicking(quizId)
				live_broadcastChans[quizId] <- models.LiveMessage{
					Type: "last_question",
					Msg:  data_to_send,
					Conn: conn}
			}

		case "next_question":

			live_mu.RLock()
			next_question_available := current_ques_indices[quizId]+1 < len(questionsPerRoom[quizId])
			timeLeftPerRoom[quizId] = quizRoom.TimerTime
			live_mu.RUnlock()

			if next_question_available {
				go startTicking(quizId)
				live_mu.Lock()
				current_ques_indices[quizId]++
				live_mu.Unlock()

				live_mu.RLock()
				last_question := current_ques_indices[quizId]+1 == len(questionsPerRoom[quizId])
				question_to_send := questionsPerRoom[quizId][current_ques_indices[quizId]]
				question_to_send.QuestionId = current_ques_indices[quizId] + 1
				data_to_send := gin.H{
					"Question": question_to_send,
					"Timer":    timeLeftPerRoom[quizId],
				}
				live_mu.RUnlock()

				if last_question {
					live_broadcastChans[quizId] <- models.LiveMessage{
						Type: "last_question",
						Msg:  data_to_send,
						Conn: conn}
				} else {
					live_broadcastChans[quizId] <- models.LiveMessage{
						Type: "question",
						Msg:  data_to_send,
						Conn: conn}
				}
			}

		case "get_scorecard":
			live_broadcastChans[quizId] <- models.LiveMessage{
				Type: "scorecard",
				Msg:  nil,
				Conn: conn}
		case "answer":
			answerData := clientMsg.Msg.(map[string]interface{})
			user_id := int(answerData["UserId"].(float64))
			answer := answerData["Answer"].(string)

			live_mu.RLock()
			question_to_send := questionsPerRoom[quizId][current_ques_indices[quizId]]
			scorecardPerRoom[quizId][user_id].CurrentAnswer = answer
			question_to_send.QuestionId = current_ques_indices[quizId] + 1
			live_mu.RUnlock()

			if answer == question_to_send.Answer {
				live_mu.Lock()
				scorecardPerRoom[quizId][user_id].CurrentScore++
				live_mu.Unlock()
			}

			live_mu.RLock()
			msg_to_send := gin.H{
				"Question":   question_to_send,
				"ScoreSheet": scorecardPerRoom[quizId],
			}
			live_mu.RUnlock()

			live_broadcastChans[quizId] <- models.LiveMessage{
				Type: "scorecard",
				Msg:  msg_to_send,
				Conn: conn}
		}
	}
}

func startTicking(quizId int) {
	tickerChanPerRoom[quizId] = time.NewTicker(1 * time.Second)
	for {
		<-tickerChanPerRoom[quizId].C
		if timeLeftPerRoom[quizId] > 0 {
			timeLeftPerRoom[quizId]--
		} else {
			tickerChanPerRoom[quizId].Stop()
			return
		}
	}
}
