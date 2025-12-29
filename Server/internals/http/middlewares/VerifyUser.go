package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/utils"
)

func VerifyMiddleware(tokenProvider auth.TokenServices) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ClientSideToken, err := req.Cookie("access_token")
			if err != nil {
				utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("Please provide a valid access token")))
				return
			}
			verifiedUser, errType, err := tokenProvider.VerifyAccessToken(req.Context(), ClientSideToken.Value)
			if err != nil {
				utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
				return
			}
			ctx := context.WithValue(req.Context(), "USER", verifiedUser)

			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}
