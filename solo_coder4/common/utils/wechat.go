package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type WechatConfig struct {
	AppID     string
	AppSecret string
}

type WechatLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func GetWechatOpenID(code string, cfg WechatConfig) (*WechatLoginResponse, error) {
	params := url.Values{}
	params.Add("appid", cfg.AppID)
	params.Add("secret", cfg.AppSecret)
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?%s", params.Encode())
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result WechatLoginResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}

	return &result, nil
}
