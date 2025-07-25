package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// initValidator initializes the validator exactly once
func initValidator() {
	once.Do(func() {
		validate = validator.New()
	})
}

func NewValidator() *validator.Validate {
	initValidator()
	return validate
}

func GetValidator() *validator.Validate {
	initValidator()
	return validate
}

func ValidateRequest(req any) error {
	initValidator()

	if err := validate.Struct(req); err != nil {
		// For now, return the validator error directly
		// Can be enhanced with custom error types later
		return err
	}
	return nil
}
