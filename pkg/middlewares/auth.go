package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"simadaservices/pkg/models"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

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

	token.Token = strings.ReplaceAll(token.Token, "Bearer ", "")
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(token.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	// if !tokenClaims.Valid {
	// 	ctx.JSON(401, "Unauthorized")
	// 	ctx.Abort()
	// 	return
	// }

	if token.Token == "" {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	marshalledJson, err := json.Marshal(token)
	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	resMsg, err := m.Nc.Request("auth.validate", marshalledJson, time.Second*10)

	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	err = json.Unmarshal(resMsg.Data, &token)

	if err != nil {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	fmt.Println(claims["org_id"], "claims")

	ctx.Set("token_info", claims)
	if !token.Response {
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	ctx.Next()
}
