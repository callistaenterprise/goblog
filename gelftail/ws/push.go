package ws

import (
	"github.com/sirupsen/logrus"
	"github.com/googollee/go-socket.io"
	"net/http"
)

var clientSocket socketio.Socket

func BroadcastLogStatement(logStatement []byte) {
	if clientSocket != nil {
		clientSocket.Emit("message", string(logStatement))
	}
}

func StartSocketIOServer() {
	var err error
	server, err := socketio.NewServer(nil)
	if err != nil {
		logrus.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		clientSocket = so
		logrus.Println("on connection")

		clientSocket.On("disconnection", func() {
			logrus.Println("on disconnect")
			clientSocket.Disconnect()
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		logrus.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	logrus.Infoln("Websocket server started at :12099...")
	logrus.Fatal(http.ListenAndServe(":12099", nil))
}
