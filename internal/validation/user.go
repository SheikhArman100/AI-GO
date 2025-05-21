package validation

import (
	"mime/multipart"
)

// IsImage validates if the provided file header is an image.
// It's exported to be registered by the main application or router setup.

type UpdateProfileRequest struct {
	Name    string                `form:"name" validate:"omitempty,min=2,max=100"`
	Address string                `form:"address" validate:"omitempty,max=255"`
	City    string                `form:"city" validate:"omitempty,max=100"`
	Road    string                `form:"road" validate:"omitempty,max=100"`
	Image   *multipart.FileHeader `form:"image" validate:"omitempty"`
}
