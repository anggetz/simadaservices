package pubsub

import (
	"encoding/json"
	"libcore/models"
	"service-auth/kernel"
	"service-auth/usecase"

	"github.com/nats-io/nats.go"
)

type AuthPubSub interface {
	RegisterAuthValidation()
	RegisterGetUser()
}

type authPubSubImpl struct {
	nc *nats.Conn
}

// RegisterGetUser implements AuthPubSub.
func (a *authPubSubImpl) RegisterGetUser() {
	a.nc.Subscribe("auth.getuser", func(msg *nats.Msg) {
		payload := struct {
			UserID      float64
			permissions []string
		}{}

		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			panic(err)
		}

		respon := usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).IsUserHasAccess(payload.UserID, payload.permissions)

		result, err := json.Marshal(struct {
			respon bool
		}{
			respon: respon,
		})

		if err != nil {
			panic(err)
		}

		msg.Respond([]byte(result))
	})
}

// RegisterAuthValidation implements AuthPubSub.
func (a *authPubSubImpl) RegisterAuthValidation() {
	a.nc.Subscribe("auth.validate", func(msg *nats.Msg) {
		token := struct {
			Token    string
			User     models.User
			Response bool
		}{}

		err := json.Unmarshal(msg.Data, &token)
		if err != nil {
			panic(err)
		}

		token.User, token.Response = usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).ValidateToken(token.Token)

		tokenMarshalled, err := json.Marshal(token)
		if err != nil {
			panic(err)
		}

		msg.Respond([]byte(tokenMarshalled))
	})
}

func NewPubSup() AuthPubSub {
	return &authPubSubImpl{}
}
