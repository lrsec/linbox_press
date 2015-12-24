package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	log "github.com/cihub/seelog"
	"linbox_stress/message"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	defer log.Flush()

	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Error("Can not read config file:", err)
		return
	}

	serverAddress := configuration.serverHost + ":" + configuration.serverIp

	for i := 0; i < configuration.threads; i++ {
		sender := configuration.senderStart + int64(i)
		receiver := configuration.receiverStart + int64(i)

		sendConn, receiveConn, closeSignal, err := createReadWritePair(serverAddress, sender, receiver)
		if err != nil {
			log.Errorf("Connect to server address fail for id pair: %d - %d\n", sender, receiver)
			continue
		}

		go sendMessage(sendConn, closeSignal, sender, receiver)
		go receiveMessage(receiveConn, closeSignal, sender, receiver)
	}

	for {

	}

}

type config struct {
	serverHost    string `json:"server_host"`
	serverIp      string `json:"server_ip"`
	senderStart   int64  `json:"sender_start"`
	receiverStart int64  `json:"receiver_start"`
	threads       int    `json:"threads"`
}

func createReadWritePair(address string, fromUserId, remoteUserId int64) (sendConn, receiveConn net.Conn, closeSignal chan bool, err error) {
	defer func() {

		if r := recover(); r != nil {
			if sendConn != nil {
				sendConn.Close()
			}
			if receiveConn != nil {
				receiveConn.Close()
			}
			if closeSignal != nil {
				close(closeSignal)
			}

			sendConn = nil
			receiveConn = nil
			closeSignal = nil
		}
	}()

	sendConn, err = createConnect(address, fromUserId, "token")
	if err != nil {
		panic(err)
	}

	receiveConn, err = createConnect(address, remoteUserId, "token")
	if err != nil {
		panic(err)
	}

	closeSignal = make(chan bool)

	return
}

func createConnect(address string, userId int64, token string) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp4", address, 6*time.Second)
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok {
				log.Errorf("Create connection fail", err)
			}

			if conn != nil {
				conn.Close()
			}

			conn = nil
		}
	}()

	localAddress := strings.Split(conn.LocalAddr().String(), ":")
	authReqeust := message.NewAuthRequest(userId, token, localAddress[0], localAddress[1], "test-env")
	content, err := message.BuildBytes(message.AUTH_REQUEST_MSG, authReqeust)
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(content)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReaderSize(conn, 100)

	typeRaw, err := reader.Peek(2)
	if err != nil {
		panic(err)
	}

	responseType, err := message.ParseRequestResponseType(uint64(binary.BigEndian.Uint16(typeRaw)))
	if responseType != message.AUTH_REQUEST_MSG {
		panic(errors.New("The first answer is not AuthResponse"))
	}

	lengthRaw, err := reader.Peek(4)
	if err != nil {
		panic(err)
	}

	length := binary.BigEndian.Uint32(lengthRaw)
	contentRaw, err := reader.Peek(int(length))
	if err != nil {
		panic(err)
	}

	response := message.AuthResponse{}
	err = json.Unmarshal(contentRaw, message.AuthResponse{})
	if err != nil {
		panic(err)
	}

	if response.Status != 200 {
		panic(errors.New("Auth Response fail: " + response.ErrMsg))
	}

	return conn, nil
}

func sendMessage(conn net.Conn, closeSignal chan bool, sender, receiver int64) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok {
				log.Errorf("send message from %d to %d panic.", sender, receiver, err)
			}

			if conn != nil {
				conn.Close()
			}

			closeSignal <- true

			close(closeSignal)
		}
	}()

}

func receiveMessage(conn net.Conn, closeSignal chan bool, sender, receiver int64) {

}
