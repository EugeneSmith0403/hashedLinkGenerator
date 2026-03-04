package auth

import (
	"adv/go-http/configs"
	internalJWT "adv/go-http/internal/jwt"
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/helpers"
	"adv/go-http/pkg/redis"
	"adv/go-http/pkg/request"
	"adv/go-http/pkg/response"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandlerDeps struct {
	*configs.Config
	AuthService *AuthService
	JWTService  *internalJWT.JWTService
	RedisSrvice *redis.Redis
}

type AuthHandler struct {
	*configs.Config
	responsePkg response.Response
	AuthService *AuthService
	JWTService  *internalJWT.JWTService
	RedisSrvice *redis.Redis
}

func NewAuthHandlers(router *http.ServeMux, deps AuthHandlerDeps) {

	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &AuthHandler{
		Config:      deps.Config,
		responsePkg: *response.NewResponse(options),
		AuthService: deps.AuthService,
		JWTService:  deps.JWTService,
		RedisSrvice: deps.RedisSrvice,
	}
	router.HandleFunc("POST /auth/login", handler.Login())
	router.HandleFunc("POST /auth/register", handler.Register())
}

func (auth *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[LoginRequest](req, w, auth.responsePkg)

		if err != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Writer: w,
				Reader: req,
				Code:   423,
			})
			return
		}

		isAuth := auth.AuthService.Login(body.Email, body.Password)

		if isAuth == false {
			auth.responsePkg.Json(&response.JsonOptions{
				Data: errorType.ErrorType{
					Error: Unauthorized,
				},
				Writer: w,
				Reader: req,
				Code:   401,
			})
			return
		}
		now := time.Now()
		expiredHours, expiredHoursErr := strconv.Atoi(auth.Config.Auth.ExpiredAt)

		if err != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data: errorType.ErrorType{
					Error: expiredHoursErr.Error(),
				},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		expirationTime := now.Add(helpers.ToHours(expiredHours)).Unix()

		claims := jwt.MapClaims{
			"email":     body.Email,
			"createdAt": now,
			"exp":       expirationTime,
			"iat":       now.Unix(),
		}

		token, tokenErr := auth.JWTService.GenerateToken(&claims)

		if tokenErr != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data: errorType.ErrorType{
					Error: tokenErr.Error(),
				},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		res := &LoginResponse{
			Token: token,
			Email: body.Email,
		}

		userKey := fmt.Sprintf("token:%s", body.Email)

		auth.RedisSrvice.Set(userKey, true, time.Duration(expirationTime))

		auth.responsePkg.Json(&response.JsonOptions{
			Data:   res,
			Writer: w,
			Reader: req,
			Code:   200,
		})
	}
}

func (auth *AuthHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[RegisterRequest](req, w, auth.responsePkg)

		if err != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Writer: w,
				Reader: req,
				Code:   423,
			})
			return
		}

		email, regError := auth.AuthService.Register(body.Name, body.Email, body.Password)

		if regError != nil {

			code := map[bool]int{true: http.StatusConflict, false: http.StatusInternalServerError}[regError.Error() == UserExists]

			auth.responsePkg.Json(&response.JsonOptions{
				Data: errorType.ErrorType{
					Error: regError.Error(),
				},
				Writer: w,
				Reader: req,
				Code:   code,
			})
			return
		}

		res := &RegisterResponse{
			LoginResponse: &LoginResponse{
				Email: email,
			},
		}

		auth.responsePkg.Json(&response.JsonOptions{
			Data:   res,
			Writer: w,
			Reader: req,
			Code:   201,
		})
	}
}
