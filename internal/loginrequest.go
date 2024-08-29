package internal

import (
	"github.com/dogclan/dumbspy/pkg/gamespy"

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

func NewGamespyLoginRequest(packet *gamespy.Packet) *GamespyLoginRequest {
	r := &GamespyLoginRequest{}
	packet.Do(func(element gamespy.KeyValuePair) {
		switch element.Key {
		case "login":
			r.Login = Pointer(element.Value)
		case "challenge":
			r.Challenge = element.Value
		case "uniquenick":
			r.UniqueNick = element.Value
		case "response":
			r.Response = element.Value
		case "port":
			r.Port = element.Value
		case "productid":
			r.ProductID = element.Value
		case "gamename":
			r.GameName = element.Value
		case "namespaceid":
			r.NamespaceID = element.Value
		case "sdkrevision":
			r.SDKRevision = element.Value
		case "id":
			r.ID = element.Value
		}
	})

	return r
}

func (r *GamespyLoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
