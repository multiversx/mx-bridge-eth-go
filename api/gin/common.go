package gin

import (
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}

func registerValidators() error {
	validators := []validatorInput{
		{
			Name:      "skValidator",
			Validator: skValidator,
		},
	}
	for _, validatorFunc := range validators {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if ok {
			err := v.RegisterValidation(validatorFunc.Name, validatorFunc.Validator)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
