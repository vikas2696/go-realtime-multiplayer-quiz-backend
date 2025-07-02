package routeshandlers

import (
	"encoding/json"
	"fmt"
	"go-multiplayer-quiz-project/backend/models"
	"go-multiplayer-quiz-project/backend/utils"
	"log"

	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	joinedPlayers  = make(map[int][]*websocket.Conn)
	mu             sync.RWMutex
	broadcastChans = make(map[int]chan models.LobbyMessage)
)

func webSocketLobby(context *gin.Context) {
	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quizroom"})
		return
	}

	var quizRoom models.QuizRoom
	err = quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		context.Writer.WriteHeader(http.StatusBadRequest)
		context.Writer.Write([]byte("Unable to upgrade connection"))
		return
	}

	_, exists := broadcastChans[quizId]
	if !exists { //host creates the room and joins and starts both goroutines which will forever listens for other players to join and update the player list
		broadcastChans[quizId] = make(chan models.LobbyMessage)
	}

	go braodcastAll(quizId)
	go readMessages(conn, quizId)

	mu.Lock()
	joinedPlayers[quizId] = append(joinedPlayers[quizId], conn)
	mu.Unlock()
	broadcastChans[quizId] <- models.LobbyMessage{Type: "join", Msg: currentPlayer, Conn: conn}
}

func braodcastAll(quizId int) { // for adding another player and then braodcast updated player list

	for { //broadcast code

		lobbyMessage := <-broadcastChans[quizId] //channel to trigger when message receives

		if lobbyMessage.Type == "leave" {

			for i, c := range joinedPlayers[quizId] {

				if c == lobbyMessage.Conn {
					mu.Lock()
					joinedPlayers[quizId] = append(joinedPlayers[quizId][:i], joinedPlayers[quizId][i+1:]...)
					mu.Unlock()
					break
				}
			}

			lobbyMessage.Conn.Close()
		}

		mu.RLock()
		conns := append([]*websocket.Conn{}, joinedPlayers[quizId]...)
		mu.RUnlock()

		for _, connection := range conns { // per room
			err := connection.WriteJSON(lobbyMessage)
			if err != nil {
				fmt.Println("websocket write error", err)
				connection.Close()
			}
		}
	}
}

func readMessages(conn *websocket.Conn, quizId int) { // to read messages from the frontend

	var lobbyClientMessage models.LobbyMessage
	for { // per connection
		_, msgJSON, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Connection is closed", err)
			//conn.Close()
			return
		}

		err = json.Unmarshal(msgJSON, &lobbyClientMessage)
		if err != nil {
			fmt.Println("Wrong message format: ", err)
		}

		lobbyClientMessage.Conn = conn
		broadcastChans[quizId] <- lobbyClientMessage //triggers broadcast channel
	}
}
