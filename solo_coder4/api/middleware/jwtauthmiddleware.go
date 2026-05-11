package middleware

import (
	"context"
	"net/http"
	"strings"

	"solo-coder4/common/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type JwtAuthMiddleware struct {
	secret string
}

func NewJwtAuthMiddleware(secret string) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{secret: secret}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.Error(w, http.ErrNoCookie)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			httpx.Error(w, http.ErrNoCookie)
			return
		}

		claims, err := utils.ParseToken(parts[1], m.secret)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", claims.UserId)
		next(w, r.WithContext(ctx))
	}
}

func GetUserIdFromCtx(ctx context.Context) (int64, bool) {
	userId, ok := ctx.Value("userId").(int64)
	return userId, ok
}
