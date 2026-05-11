package types

type RegisterRequest struct {
	Username string `json:"username" example:"testuser"`
	Password string `json:"password" example:"123456"`
	Email    string `json:"email" example:"test@example.com"`
	Phone    string `json:"phone" example:"13800138000"`
}

type RegisterResponse struct {
	UserId    int64  `json:"userId" example:"1"`
	Username  string `json:"username" example:"testuser"`
	LoginType int    `json:"loginType" example:"1"`
	Message   string `json:"message" example:"Registration successful"`
}

type LoginRequest struct {
	LoginType int    `json:"loginType" example:"1" enums:"1,2"`
	Username  string `json:"username,omitempty" example:"testuser"`
	Password  string `json:"password,omitempty" example:"123456"`
	Code      string `json:"code,omitempty" example:"wx_code_xxx"`
}

type LoginResponse struct {
	UserId    int64  `json:"userId" example:"1"`
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Username  string `json:"username" example:"testuser"`
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"`
	LoginType int    `json:"loginType" example:"1"`
}

type UserInfoResponse struct {
	UserId    int64  `json:"userId" example:"1"`
	Username  string `json:"username" example:"testuser"`
	Email     string `json:"email" example:"test@example.com"`
	Phone     string `json:"phone" example:"13800138000"`
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"`
	LoginType int    `json:"loginType" example:"1"`
}

const (
	LoginTypeAccount = 1
	LoginTypeWechat  = 2
)
