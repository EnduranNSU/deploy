package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type StartResetRequest struct {
	Email string `json:"email"`
}

type ConfirmResetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// type ValidateResponse struct {
// 	Sub       string         `json:"sub"`
// 	Issuer    string         `json:"iss"`
// 	ExpiresAt time.Time      `json:"exp_at"`
// 	Claims    map[string]any `json:"claims"`
// }

type ValidateResponse struct {
	UserID string `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StartResetDevResponse struct {
	DevCode string `json:"dev_code"`
}
