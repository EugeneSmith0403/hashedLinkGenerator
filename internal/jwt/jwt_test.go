package jwt_test

import (
	internalJwt "adv/go-http/internal/jwt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJwt(t *testing.T) {
	const password = "$2a$10$rPoAoGvsaQZ/tRWf3DZphuUs1LbWIky0XppainCMrISRcDe8FOH0C"
	jwtDeps := &internalJwt.JwtDeps{
		Secret: "sdfsdfs",
	}
	mapClaims := &jwt.MapClaims{"password": password}

	jwtService := internalJwt.NewJWTService(*jwtDeps)

	token, err := jwtService.GenerateToken(mapClaims)

	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatalf("Expect token")
	}

	checkedClaims, er := jwtService.CheckToken(token, mapClaims)

	if er != nil {
		t.Fatal(err)
	}

	if (*checkedClaims)["password"] == "" {
		t.Fatalf("Password not equil %s, ecpected %s", (*checkedClaims)["password"], password)
	}

}
