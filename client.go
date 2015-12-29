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
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	defer log.Flush()

	logger, err := log.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		log.Errorf("load see log config fail", err)
		return
	}
	log.ReplaceLogger(logger)

	//file, err := os.Open("/Users/lrsec/Code/im_stress/bin/config.json")
	file, err := os.Open("config.json")
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

	monitor := newTimeMonitor()

	go func() {
		for i := 0; i < configuration.Threads; i++ {
			sender := configuration.SenderStart + int64(i)
			receiver := configuration.ReceiverStart + int64(i)

			sendConn, sendCoder, err := createConnect(serverAddress, sender, "token")
			if err != nil {
				log.Errorf("Create connection for %d fail. ", sender, err)

				continue
			}

			closeSignal := make(chan bool)

			go sendMessage(sendConn, sendCoder, closeSignal, sender, receiver, monitor)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			times := monitor.checkout()
			var sum int64
			var avg float64
			length := len(times)
			if length > 0 {
				for _, t := range times {
					sum += t
				}

				avg = float64(sum) / float64(length)
			}

			log.Infof("Monitor: total message count %d. average latency: %f", length, avg)
		}
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

	conn.SetDeadline(time.Now().Add(6 * time.Second))
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

	responseType := message.RequestResponseType(binary.BigEndian.Uint16(typeRaw))
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

	length := binary.BigEndian.Uint32(lengthRaw)

	// content
	conn.SetDeadline(time.Now().Add(6 * time.Second))
	contentRaw, err := reader.Peek(int(length))
	if err != nil {
		panic(err)
	}
	reader.Discard(int(length))

	response := message.AuthResponse{}
	err = coder.Decode(contentRaw, &response)
	if err != nil {
		panic(err)
	}

	if response.Status != 200 {
		panic(errors.New("Auth Response fail: " + response.ErrMsg))
	}

	return conn, coder, nil
}

func sendMessage(conn net.Conn, coder *codec.MsgCodec, closeSignal chan bool, sender, receiver int64, monitor *TimeMonitor) {
	defer func() {
		if r := recover(); r != nil {

			log.Error("Send messaage fail. ", r)

			if conn != nil {
				conn.Close()
			}

			closeSignal <- true

			close(closeSignal)
		}
	}()

	text := "test message"
	reader := bufio.NewReaderSize(conn, 4096)

	for {
		conn.SetDeadline(time.Time{})

		requestPtr := createSendMsgRequest(sender, receiver, text)

		content, err := coder.Encode(message.SEND_MSG_REQUEST_MSG, requestPtr)
		if err != nil {
			panic(err)
		}

		startTime := time.Now().UnixNano() / int64(1000000)

		_, err = conn.Write(content)
		if err != nil {
			panic(err)
		}

		// version 信息,不关心
		_, err = reader.Discard(2)
		if err != nil {
			panic(err)
		}

		// Type
		typeRaw, err := reader.Peek(2)
		if err != nil {
			panic(err)
		}
		reader.Discard(2)

		responseType := message.RequestResponseType(binary.BigEndian.Uint16(typeRaw))
		if responseType != message.SEND_MSG_RESPONSE_MSG {
			panic(errors.New("Received answer is not SEND_MSG_RESPONSE_MSG"))
		}

		// length
		lengthRaw, err := reader.Peek(4)
		if err != nil {
			panic(err)
		}
		reader.Discard(4)

		length := binary.BigEndian.Uint32(lengthRaw)

		// content
		contentRaw, err := reader.Peek(int(length))
		if err != nil {
			panic(err)
		}
		reader.Discard(int(length))

		endTime := time.Now().UnixNano() / int64(1000000)

		response := message.SendMsgResponse{}
		err = coder.Decode(contentRaw, &response)
		if err != nil {
			panic(err)
		}

		if response.Status != 200 {
			panic(errors.New("Send Message Response fail: " + response.ErrMsg))
		} else {
			monitor.submit(endTime - startTime)
		}
	}
}

func createSendMsgRequest(sender, receiver int64, content string) *message.SendMsgRequest {
	rid := time.Now().UnixNano() / int64(1000000)
	fromUserId := strconv.FormatInt(sender, 10)
	toUserId := strconv.FormatInt(receiver, 10)
	mimeType := "text/plain"

	msg := message.Message{}
	msg.RId = rid
	msg.FromUserId = fromUserId
	msg.ToUserId = toUserId
	msg.MimeType = mimeType
	msg.Content = content
	msg.SendTime = rid
	msg.Type = int(message.MESSAGE_TYPE_SESSION)

	request := &message.SendMsgRequest{}
	request.RId = rid
	request.UserId = fromUserId
	request.RemoteId = toUserId
	request.Msg = msg
	request.Type = int(message.MESSAGE_TYPE_SESSION)

	return request
}

const default_monitor_buffer int = 100000

type TimeMonitor struct {
	times  []int64
	locker sync.Locker
}

func newTimeMonitor() *TimeMonitor {
	tm := &TimeMonitor{}
	tm.times = make([]int64, 0, default_monitor_buffer)
	tm.locker = &sync.Mutex{}

	return tm
}

func (tm *TimeMonitor) submit(time int64) {
	tm.locker.Lock()
	defer tm.locker.Unlock()

	tm.times = append(tm.times, time)
}

func (tm *TimeMonitor) checkout() []int64 {
	tm.locker.Lock()
	defer tm.locker.Unlock()

	var t []int64
	t, tm.times = tm.times, make([]int64, 0, default_monitor_buffer)

	return t
}
