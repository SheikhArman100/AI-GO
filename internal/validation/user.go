package validation

import "mime/multipart"

// UpdateProfileRequest defines the validation rules for updating a user's profile.
// Image   string `json:"image"` // Assuming image update is handled separately or via a URL for now

// We can add validation tags here, e.g., `validate:"required,min=3"`.
type UpdateProfileRequest struct {
	Name    string `form:"name" validate:"omitempty,min=2,max=100"`
	Address string `form:"address" validate:"omitempty,max=255"`
	City    string `form:"city" validate:"omitempty,max=100"`
	Road    string `form:"road" validate:"omitempty,max=100"`
	Image   *multipart.FileHeader `form:"image"` // For file uploads, ensure your handler can process multipart/form-data
}
