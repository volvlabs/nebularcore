package validation

import (
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

// defaultPhoneRegion is the region used by the "phonenumber" struct-tag
// validator when constructed via New() — preserved for backward
// compatibility with every existing single-region caller. Multi-country
// apps should use NewWithPhoneRegion instead, resolving the region
// per-instance (e.g. from a request's resolved account/detected country)
// rather than relying on this hardcoded default.
const defaultPhoneRegion = "NG"

func New() *Validator {
	return NewWithPhoneRegion(defaultPhoneRegion)
}

// NewWithPhoneRegion builds a Validator whose "phonenumber" struct-tag
// validator parses numbers against phoneRegion (a libphonenumber region
// code, e.g. "GH") instead of the hardcoded "NG" default. The underlying
// ValidatePhoneNumber function has always been region-aware — this was the
// only call site that hardcoded a region.
//
// A single *validator.Validate instance only supports one registered
// "phonenumber" region at a time, so a multi-country app validating phone
// numbers for users in different countries within the same process should
// construct one Validator per resolved region as needed (e.g. per-request,
// keyed by the caller's resolved country), rather than relying on a single
// process-wide default.
func NewWithPhoneRegion(phoneRegion string) *Validator {
	validate := validator.New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic(err)
	}

	if err := validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	}); err != nil {
		panic(err)
	}

	v := &Validator{
		validate:   validate,
		translator: trans,
	}

	if err := validate.RegisterValidation("phonenumber", func(fl validator.FieldLevel) bool {
		return ValidatePhoneNumber(fl.Field().String(), phoneRegion)
	}); err != nil {
		panic(err)
	}

	if err := validate.RegisterValidation("custom_email", func(fl validator.FieldLevel) bool {
		isValid, _ := ValidateEmail(fl.Field().String())
		return isValid
	}); err != nil {
		panic(err)
	}

	return v
}

func (v *Validator) GetValidate() *validator.Validate {
	return v.validate
}

func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err != nil {
		errs := []types.FieldError{}
		for _, err := range err.(validator.ValidationErrors) {
			errs = append(errs, types.FieldError{
				Field:   err.Field(),
				Message: err.Translate(v.translator),
			})
		}
		return types.NewValidationError("Validation failed. Please check the provided values and try again.", errs)
	}

	return nil
}
