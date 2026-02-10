package types

import "github.com/golang-jwt/jwt/v5"

type UserCredential struct {
	Site  string `json:"site"`
	Email string `json:"email"`
	PWD   string `json:"pwd"`
}

type SiteData struct {
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoUrl    string `json:"photo_url"`
}

type UserData struct {
	Name     string `json:"name"`
	PhotoUrl string `json:"photo_url"`
	Email    string `json:"email"`
}

type CustomClaims struct {
	UserID   int      `json:"user_id"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	User     UserData `json:"user"`
	jwt.RegisteredClaims
}
