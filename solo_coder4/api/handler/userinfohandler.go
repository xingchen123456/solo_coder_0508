package handler

import (
	"errors"
	"net/http"

	"solo-coder4/api/logic"
	"solo-coder4/api/middleware"
	"solo-coder4/api/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户信息
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.UserInfoResponse
// @Failure 401 {object} map[string]string
// @Router /user/info [get]
func UserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := middleware.GetUserIdFromCtx(r.Context())
		if !ok {
			httpx.Error(w, errors.New("invalid token"))
			return
		}

		l := logic.NewUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetUserInfo(userId)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
