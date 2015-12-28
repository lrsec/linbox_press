package message

type SendMsgRequest struct {
	RId      int64   `json:"r_id"`
	UserId   string  `json:"user_id"`
	RemoteId string  `json:"remote_id"`
	GroupId  string  `json:"group_id"`
	Msg      Message `json:"msg"`
	Type     int     `json:"type"`
}

type Message struct {
	RId        int64  `json:"r_id"`
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
	GroupId    string `json:"group_id"`
	MsgId      int64  `json:"msg_id"`
	MimeType   string `json:"mime_type"`
	Content    string `json:"content"`
	SendTime   int64  `json:"send_time"`
	Type       int    `json:"type"`
}
