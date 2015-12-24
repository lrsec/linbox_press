package message

type AuthResponse struct {
	RId      int64  `json:"r_id"`
	UserId   int64  `json:"user_id"`
	Status   int32  `json:"status"`
	ErrMsg   string `json:"err_msg"`
	SendTime int64  `json:"send_time"`
}
