package dto

// RegisterRequest — запрос на регистрацию.
// @Description Данные для регистрации нового пользователя.
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginRequest — запрос на аутентификацию.
// @Description Данные для входа.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RefreshRequest — запрос на обновление токенов.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse — ответ с парой токенов.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
