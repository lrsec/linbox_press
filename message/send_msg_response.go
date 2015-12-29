package message

type SendMsgResponse struct {
	RId      int64  `json:"r_id"`
	MsgRId   int64  `json:"msg_r_id"`
	UserId   string `json:"user_id"`
	RemoteId string `json:"remote_id"`
	GroupId  string `json:"group_id"`
	MsgId    int64  `json:"msg_id"`
	SendTime int64  `json:"send_time"`
	Type     int    `json:"type"`
	Status   int32  `json:"status"`
	ErrMsg   string `json:"err_msg"`
}
