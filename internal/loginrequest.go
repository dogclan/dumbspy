package internal

import (
	"dogclan/dumbspy/pkg/packet"

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

func NewGamespyLoginRequest(packet *packet.GamespyPacket) *GamespyLoginRequest {
	data := packet.Map()
	return &GamespyLoginRequest{
		Login:       Pointer[string](data["login"]),
		Challenge:   data["challenge"],
		UniqueNick:  data["uniquenick"],
		Response:    data["response"],
		Port:        data["port"],
		ProductID:   data["productid"],
		GameName:    data["gamename"],
		NamespaceID: data["namespaceid"],
		SDKRevision: data["sdkrevision"],
		ID:          data["id"],
	}
}

func (r *GamespyLoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
