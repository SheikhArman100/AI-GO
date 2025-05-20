package validation

// SignUpRequest defines the validation schema for sign-up
type SignUpRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Email       string `json:"email" binding:"required,email,max=255"`
	PhoneNumber string `json:"phone_number" binding:"required,max=20"`
	Password    string `json:"password" binding:"required,min=6"`
}
// SignInRequest defines the validation schema for sign-in
type SignInRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6"`
}
