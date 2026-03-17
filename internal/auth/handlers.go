package auth

import (
	authsession "link-generator/internal/auth_session"
	"link-generator/internal/models"
	errorType "link-generator/pkg/errorType"
	"link-generator/pkg/helpers"
	"link-generator/pkg/limiter"
	"link-generator/pkg/middleware"
	"link-generator/pkg/request"
	"link-generator/pkg/response"
	"net/http"
)

type AuthHandlerDeps struct {
	AuthService        *AuthService
	AuthMailerDeps     AuthMailerDeps
	AccountService     models.IAccountService
	AuthSeseionService *authsession.AuthSessionService
	IPRateLimiter      *limiter.LimiterService
	UserRepository     IUserRepository
}

type AuthHandler struct {
	responsePkg        response.Response
	AuthService        *AuthService
	authMailer         *AuthMailer
	AccountService     models.IAccountService
	AuthSeseionService *authsession.AuthSessionService
	UserRepository     IUserRepository
}

func NewAuthHandlers(router *http.ServeMux, deps AuthHandlerDeps) {

	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &AuthHandler{
		responsePkg:        *response.NewResponse(options),
		AuthService:        deps.AuthService,
		authMailer:         NewAuthMailer(deps.AuthMailerDeps),
		AccountService:     deps.AccountService,
		AuthSeseionService: deps.AuthSeseionService,
		UserRepository:     deps.UserRepository,
	}
	ipRateLimit := middleware.RateLimit(deps.IPRateLimiter, limiter.KeyByIP)

	router.Handle("POST /auth/login", ipRateLimit(handler.Login()))
	router.Handle("POST /auth/register", ipRateLimit(handler.Register()))
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
			generatedToken, expTime, tokenErr := auth.AuthService.GenerateToken(body.Email)
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

			_, sessionErr := auth.AuthSeseionService.Update(&authsession.AddOptions{
				AccountID: accInfo.AccountID,
				Token:     token,
				ExpiresAt: expTime,
				IsVerify:  true,
				IpAddress: helpers.GetClientIP(req),
				UserAgent: req.UserAgent(),
			})
			if sessionErr != nil {
				auth.responsePkg.Json(&response.JsonOptions{
					Data:   errorType.ErrorType{Error: sessionErr.Error()},
					Writer: w,
					Reader: req,
					Code:   http.StatusInternalServerError,
				})
				return
			}
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

		foundUser, err := auth.UserRepository.FindByEmail(email)
		if err != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		if foundUser == nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "user not found"},
				Writer: w,
				Reader: req,
				Code:   http.StatusNotFound,
			})
			return
		}

		account, createErr := auth.AccountService.CreateAccount(foundUser.Model.ID, foundUser.Name, foundUser.Email)
		if createErr != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: createErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		token, expTime, tokenErr := auth.AuthService.GenerateToken(email)
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

		_, upErr := auth.AuthSeseionService.Update(&authsession.AddOptions{
			AccountID: account.Model.ID,
			Token:     token,
			ExpiresAt: expTime,
			IsVerify:  false,
			IpAddress: helpers.GetClientIP(req),
			UserAgent: req.UserAgent(),
		})

		if upErr != nil {
			auth.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: upErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
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
