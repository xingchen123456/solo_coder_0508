package types

type LoginType int

const (
	LoginTypeAccount LoginType = iota
	LoginTypeWechat
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	LoginType int    `json:"loginType"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Code      string `json:"code,omitempty"`
}

type LoginResponse struct {
	UserId   int64  `json:"userId"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type UserInfo struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}
