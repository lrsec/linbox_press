package message

import "time"

type AuthRequest struct {
	RId    int64  `json:"r_id"`
	UserId int64  `json:"user_id"`
	Token  string `json:token`
	Ip     string `json:"ip"`
	Port   string `json:"port"`
	Device string `json:"device"`
}

func NewAuthRequest(userId int64, token, ip, port, device string) *AuthRequest {
	request := &AuthRequest{
		RId:    time.Now().UnixNano() / 1000000,
		UserId: userId,
		Token:  token,
		Ip:     ip,
		Port:   port,
		Device: device,
	}

	return request
}
