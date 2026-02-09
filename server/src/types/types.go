package types

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
