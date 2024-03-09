package middlewares

import (
	"encoding/json"
	"fmt"
	"libcore/models"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type middlewareAuth struct {
	Nc     *nats.Conn
	jwtKey string
}

func NewMiddlewareAuth(nc *nats.Conn) *middlewareAuth {
	return &middlewareAuth{
		Nc: nc,
	}
}

func (m *middlewareAuth) SetJwtKey(jwtKey string) *middlewareAuth {
	m.jwtKey = jwtKey
	return m
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
		return []byte(m.jwtKey), nil
	})

	// if !tokenClaims.Valid {
	// 	ctx.JSON(401, "Unauthorized")
	// 	ctx.Abort()
	// 	return
	// }

	if token.Token == "" {
		fmt.Println("error token is nil")
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	marshalledJson, err := json.Marshal(token)
	if err != nil {
		fmt.Println("error marshalled token", err.Error())
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	resMsg, err := m.Nc.Request("auth.validate", marshalledJson, time.Second*10)

	if err != nil {
		fmt.Println("error request validate token", err.Error())
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	err = json.Unmarshal(resMsg.Data, &token)

	if err != nil {
		fmt.Println("error unmarshall token", err.Error())
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	ctx.Set("token_info", claims)
	if !token.Response {
		fmt.Println("error get claim token", token)
		ctx.JSON(401, "Unauthorized")
		ctx.Abort()
		return
	}

	ctx.Next()
}
