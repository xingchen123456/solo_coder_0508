package handler

import (
	"net/http"

	"solo-coder4/api/logic"
	"solo-coder4/api/svc"
	"solo-coder4/api/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// @Summary 用户注册
// @Description 用户通过用户名和密码注册新账号
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body types.RegisterRequest true "注册信息"
// @Success 200 {object} types.RegisterResponse
// @Failure 400 {object} map[string]string
// @Router /register [post]
func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
