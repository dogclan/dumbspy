package internal

import (
	"github.com/go-playground/validator/v10"
)

type GamespyLoginRequest struct {
	Login       *string `validate:"len=0,required"` // Key only, must be empty
	Challenge   string  `validate:"len=32,required"`
	UniqueNick  string  `validate:"min=1,required"`
	Response    string  `validate:"md5,required"`
	Port        string  `validate:"numeric,required"`
	ProductID   string  `validate:"numeric,required"`
	GameName    string  `validate:"min=1,required"`
	NamespaceID string  `validate:"numeric,required"`
	SDKRevision string  `validate:"numeric,required"`
	ID          string  `validate:"numeric,required"`
}

func (r GamespyLoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
