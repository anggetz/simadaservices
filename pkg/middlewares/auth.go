package middlewares

import (
	"encoding/json"
	"simadaservices/pkg/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type middlewareAuth struct {
	Nc *nats.Conn
}

func NewMiddlewareAuth(nc *nats.Conn) *middlewareAuth {
	return &middlewareAuth{
		Nc: nc,
	}
}

func (m *middlewareAuth) TokenValidate(ctx *gin.Context) {
	token := struct {
		Token    string
		User     *models.User
		Response bool
	}{
		Token:    ctx.Request.Header.Get("Authorization"),
		Response: true,
	}

	marshalledJson, err := json.Marshal(token)
	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
	}

	resMsg, err := m.Nc.Request("auth.validate", marshalledJson, time.Second*10)

	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
	}

	err = json.Unmarshal(resMsg.Data, &token)

	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
	}

	ctx.Set("user_logged_in_username", token.User.Username)
	ctx.Set("user_logged_in_org_id", token.User.PidOrganisasi)
	if !token.Response {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
	}

	ctx.Next()
}
