package gema

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type jsonSerializer struct{}

func (d jsonSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}

	return enc.Encode(i)
}

func (d jsonSerializer) Deserialize(c echo.Context, i interface{}) error {
	err := json.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}

	return err
}

var decorateJsonSerializer = fx.Decorate(func(e *echo.Echo) *echo.Echo {
	e.JSONSerializer = &jsonSerializer{}
	return e
})
