package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	log "github.com/cihub/seelog"
	"linbox_stress/codec"
	"linbox_stress/message"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	defer log.Flush()

	file, err := os.Open("/Users/lrsec/Work/medtree/linbox_stress/bin/config.json")
	//file, err := os.Open("config.json")
	if err != nil {
		log.Error("Can not open config file", err)
		return
	}

	decoder := json.NewDecoder(file)
	configuration := config{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Error("Can not read config file:", err)
		return
	}

	serverAddress := configuration.ServerIp + ":" + configuration.ServerPort

	log.Info("Server address from config: ", serverAddress)

	for i := 0; i < configuration.Threads; i++ {
		sender := configuration.SenderStart + int64(i)
		receiver := configuration.ReceiverStart + int64(i)

		sendConn, sendCoder, err := createConnect(serverAddress, sender, "token")
		if err != nil {
			log.Errorf("Create connection for %d fail. ", sender, err)

			continue
		}

		receiveConn, receiveCoder, err := createConnect(serverAddress, receiver, "token")
		if err != nil {
			log.Errorf("Create connection for %d fail. ", receiver, err)

			sendConn.Close()
			continue
		}

		closeSignal := make(chan bool)

		go sendMessage(sendConn, sendCoder, closeSignal, sender, receiver)
		go receiveMessage(receiveConn, receiveCoder, closeSignal, sender, receiver)
	}

	for {

	}

}

type config struct {
	ServerIp      string `json:"server_ip"`
	ServerPort    string `json:"server_port"`
	SenderStart   int64  `json:"sender_start"`
	ReceiverStart int64  `json:"receiver_start"`
	Threads       int    `json:"threads"`
}

func createConnect(address string, userId int64, token string) (conn net.Conn, coder *codec.MsgCodec, err error) {
	conn, err = net.DialTimeout("tcp4", address, 6*time.Second)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if r := recover(); r != nil {

			switch i := r.(type) {
			case string, error:
				log.Error("Create connection fail ", i)
			default:
				log.Error("Create connection fail ")
			}

			if conn != nil {
				conn.Close()
			}

			conn = nil
		} else {
			conn.SetDeadline(time.Time{})
		}
	}()

	coder, err = codec.NewMsgCodec()
	if err != nil {
		panic(err)
	}

	localAddress := strings.Split(conn.LocalAddr().String(), ":")
	authReqeust := message.NewAuthRequest(userId, token, localAddress[0], localAddress[1], "test-env")
	content, err := coder.Encode(message.AUTH_REQUEST_MSG, authReqeust)
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(content)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReaderSize(conn, 4096)

	// version 信息,不关心
	conn.SetDeadline(time.Now().Add(6 * time.Second))
	_, err = reader.Discard(2)
	if err != nil {
		panic(err)
	}

	// Type
	conn.SetDeadline(time.Now().Add(6 * time.Second))
	typeRaw, err := reader.Peek(2)
	if err != nil {
		panic(err)
	}
	reader.Discard(2)

	log.Info("type: ", typeRaw)

	responseType, err := message.ParseRequestResponseType(uint64(binary.BigEndian.Uint16(typeRaw)))
	if responseType != message.AUTH_RESPONSE_MSG {
		panic(errors.New("The first answer is not AuthResponse"))
	}

	// length
	conn.SetDeadline(time.Now().Add(6 * time.Second))
	lengthRaw, err := reader.Peek(4)
	if err != nil {
		panic(err)
	}
	reader.Discard(4)

	log.Info("length: ", lengthRaw)

	length := binary.BigEndian.Uint32(lengthRaw)

	// content
	conn.SetDeadline(time.Now().Add(6 * time.Second))
	contentRaw, err := reader.Peek(int(length))
	if err != nil {
		panic(err)
	}
	reader.Discard(int(length))

	response := message.AuthResponse{}
	err = json.Unmarshal(contentRaw, message.AuthResponse{})
	if err != nil {
		panic(err)
	}

	if response.Status != 200 {
		panic(errors.New("Auth Response fail: " + response.ErrMsg))
	}

	return conn, coder, nil
}

func sendMessage(conn net.Conn, coder *codec.MsgCodec, closeSignal chan bool, sender, receiver int64) {
	defer func() {
		if r := recover(); r != nil {
			switch i := r.(type) {
			case string, error:
				log.Error("Create connection fail", i)
			default:
				log.Error("Create connection fail")
			}

			if conn != nil {
				conn.Close()
			}

			closeSignal <- true

			close(closeSignal)
		}
	}()

}

func receiveMessage(conn net.Conn, coder *codec.MsgCodec, closeSignal chan bool, sender, receiver int64) {

}
