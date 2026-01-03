package internal

import (
	"github.com/go-playground/validator/v10"
)

type GamespyLoginRequest struct {
	Login       *string `gamespy:"login" validate:"len=0,required"` // Key only, must be empty
	Challenge   string  `gamespy:"challenge" validate:"len=32,required"`
	UniqueNick  string  `gamespy:"uniquenick" validate:"min=1,required"`
	Response    string  `gamespy:"response" validate:"md5,required"`
	Port        string  `gamespy:"port" validate:"numeric,required"`
	ProductID   string  `gamespy:"productid" validate:"numeric,required"`
	GameName    string  `gamespy:"gamename" validate:"min=1,required"`
	NamespaceID string  `gamespy:"namespaceid" validate:"numeric,required"`
	SDKRevision string  `gamespy:"sdkrevision" validate:"numeric,required"`
	ID          string  `gamespy:"id" validate:"numeric,required"`
}

func (r GamespyLoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
