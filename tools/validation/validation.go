package validation

import (
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/volvlabs/nebularcore/tools/types"
)

type Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

func New() *Validator {
	validate := validator.New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic(fmt.Errorf("register default translations: %w", err))
	}

	if err := validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	}); err != nil {
		panic(fmt.Errorf("register required translation: %w", err))
	}

	v := &Validator{
		validate:   validate,
		translator: trans,
	}

	if err := validate.RegisterValidation("phonenumber", func(fl validator.FieldLevel) bool {
		return ValidatePhoneNumber(fl.Field().String(), "NG")
	}); err != nil {
		panic(fmt.Errorf("register phonenumber validation: %w", err))
	}

	if err := validate.RegisterValidation("custom_email", func(fl validator.FieldLevel) bool {
		isValid, _ := ValidateEmail(fl.Field().String())
		return isValid
	}); err != nil {
		panic(fmt.Errorf("register custom_email validation: %w", err))
	}

	return v
}

func (v *Validator) GetValidate() *validator.Validate {
	return v.validate
}

func (v *Validator) Validate(i any) ([]types.FieldError, error) {
	err := v.validate.Struct(i)
	if err != nil {
		errs := []types.FieldError{}
		for _, err := range err.(validator.ValidationErrors) {
			errs = append(errs, types.FieldError{
				Field:   err.Field(),
				Message: err.Translate(v.translator),
			})
		}
		return errs, err
	}

	return nil, nil
}
