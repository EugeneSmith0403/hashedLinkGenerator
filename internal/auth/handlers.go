package auth

import (
	"link-generator/internal/models"
	errorType "link-generator/pkg/errorType"
	"link-generator/pkg/request"
	"link-generator/pkg/response"
	"net/http"
)

type AuthHandlerDeps struct {
	AuthService    *AuthService
	AuthMailerDeps AuthMailerDeps
	AccountService models.IAccountService
}

type AuthHandler struct {
	responsePkg    response.Response
	AuthService    *AuthService
	authMailer     *AuthMailer
	AccountService models.IAccountService
}

func NewAuthHandlers(router *http.ServeMux, deps AuthHandlerDeps) {

	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &AuthHandler{
		responsePkg:    *response.NewResponse(options),
		AuthService:    deps.AuthService,
		authMailer:     NewAuthMailer(deps.AuthMailerDeps),
		AccountService: deps.AccountService,
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

		accInfo, accErr := auth.AccountService.GetAccountInfoByEmail(body.Email)
		if accErr != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: accErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		var token string

		if !accInfo.Is2FAEnabled {
			generatedToken, tokenErr := auth.AuthService.GenerateToken(body.Email)
			if tokenErr != nil {
				auth.responsePkg.Json(&response.JsonOptions{
					Data:   errorType.ErrorType{Error: tokenErr.Error()},
					Writer: w,
					Reader: req,
					Code:   http.StatusInternalServerError,
				})
				return
			}
			token = generatedToken
		}

		res := &LoginResponse{
			Token:        token,
			Is2FAEnabled: accInfo.Is2FAEnabled,
			Email:        body.Email,
		}

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

		token, tokenErr := auth.AuthService.GenerateToken(email)
		if tokenErr != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: tokenErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		res := &RegisterResponse{
			Email: email,
			Token: token,
		}

		go auth.authMailer.SendWelcomeEmail(body.Name, email, "en")

		auth.responsePkg.Json(&response.JsonOptions{
			Data:   res,
			Writer: w,
			Reader: req,
			Code:   201,
		})
	}
}
