package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"formify/server/internal/user"
)

func TestService_GenerateJWT(t *testing.T) {
	svc := NewService("test-secret")
	u := &user.User{
		ID:    12,
		Email: "test@user.user",
		Name:  "test",
	}

	tokenString, err := svc.GenerateJWT(u)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Fatalf("token should be valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims should be map")
	}

	assertUserClaims(t, claims, u)
	assertTimeClaims(t, claims)
}

func assertUserClaims(t *testing.T, claims jwt.MapClaims, u *user.User) {
	t.Helper()

	if got, ok := claims["user_id"].(float64); !ok || int32(got) != u.ID {
		t.Fatalf("unexpected user_id claim: %#v", claims["user_id"])
	}
	if got, _ := claims["email"].(string); got != u.Email {
		t.Fatalf("unexpected email claim: %#v", claims["email"])
	}
	if got, _ := claims["name"].(string); got != u.Name {
		t.Fatalf("unexpected name claim: %#v", claims["name"])
	}
}

func assertTimeClaims(t *testing.T, claims jwt.MapClaims) {
	t.Helper()

	if _, ok := claims["exp"].(float64); !ok {
		t.Fatalf("exp claim missing or invalid type: %#v", claims["exp"])
	}
	if _, ok := claims["iat"].(float64); !ok {
		t.Fatalf("iat claim missing or invalid type: %#v", claims["iat"])
	}
}
