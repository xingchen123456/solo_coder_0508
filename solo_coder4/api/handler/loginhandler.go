package handler

import (
	"net/http"

	"solo-coder4/api/logic"
	"solo-coder4/api/svc"
	"solo-coder4/api/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// @Summary 用户登录
// @Description 支持账号密码登录和微信登录
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body types.LoginRequest true "登录信息"
// @Success 200 {object} types.LoginResponse
// @Failure 400 {object} map[string]string
// @Router /login [post]
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
