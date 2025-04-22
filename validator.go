package gema

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
)

// to cache the validator (recommended by the docs)
var uni *ut.UniversalTranslator
var validate = validator.New(validator.WithRequiredStructEnabled())
var trans ut.Translator

type Validator interface {
	Validate() error
}

var validatorBaseType = reflect.TypeFor[Validate]()

// Validate is a struct to be embedded in your own struct to provide the validation
// functionality. It must be placed in the first field in your struct. This assumption
// is made to avoid the need to check if the struct embeds `gema.Validate` or not.
// Otherwise, you're assumed to use your own validator by implementing the `gema.Validator` interface.
type Validate struct {
	target interface{}
}

func newValidator(target interface{}) *Validate {
	return &Validate{target}
}

func (v *Validate) Validate() error {
	return validate.Struct(v.target)
}

func translate(err error) string {
	if translable, ok := err.(validator.ValidationErrors); ok {
		for _, err := range translable {
			return err.Translate(trans)
		}
	}

	return err.Error()
}

type binder struct {
	echo.DefaultBinder
}

func (b *binder) Bind(i interface{}, c echo.Context) error {
	if err := b.DefaultBinder.Bind(i, c); err != nil {
		return err
	}

	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if the value embeds `gema.Validate` in the first field.
	// If yes, replace the value with `gema.Validate`
	if val.Kind() == reflect.Struct && val.NumField() > 0 {
		firstFieldType := val.Field(0).Type()
		if firstFieldType == validatorBaseType {
			i = newValidator(i)
		}
	}

	v, ok := i.(Validator)
	if !ok {
		return nil
	}

	if err := v.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, translate(err))
	}

	return nil
}

func registerCustomBinder(e *echo.Echo) {
	e.Binder = &binder{}
}

func RegisterValidator(registry map[string]validator.Func) {
	for name, fn := range registry {
		if err := validate.RegisterValidation(name, fn); err != nil {
			panic(err)
		}
		fmt.Printf("[Gema] Custom validator registered: %s\n", name)
	}
}

func init() {
	en := en.New()
	uni = ut.New(en, en)
	trans, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
}
