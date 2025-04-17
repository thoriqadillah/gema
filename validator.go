package gema

import (
	"fmt"
	"net/http"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// to cache the validator (recommended by the docs)
var uni *ut.UniversalTranslator
var validate = validator.New(validator.WithRequiredStructEnabled())
var trans ut.Translator

type Validator interface {
	Validate() error
}

func ValidateStruct(i interface{}) error {
	return validate.Struct(i)
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

	v, ok := i.(Validator)
	if !ok {
		return nil
	}

	if err := v.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, translate(err))
	}

	return nil
}

var decorateBinder = fx.Decorate(func(e *echo.Echo) *echo.Echo {
	e.Binder = &binder{}
	return e
})

func registerValidator(registry map[string]validator.Func) fx.Option {
	return fx.Invoke(func() {
		for name, fn := range registry {
			fmt.Printf("[Gema] Registering custom %s validator\n", name)
			if err := validate.RegisterValidation(name, fn); err != nil {
				panic(err)
			}
		}
	})
}

func init() {
	en := en.New()
	uni = ut.New(en, en)
	trans, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
}
