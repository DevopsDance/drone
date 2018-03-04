package server

import (
	"net/http"
	"time"

	"github.com/drone/drone/shared/token"
	"github.com/drone/drone/store"
	"github.com/gin-gonic/gin"
)

// apitoken endpoint returns the requested user and a token signed by user.Hash
func BackstageUserApiToken(c *gin.Context) {
	u := c.Params.ByName("user")

	user, err := store.GetUserLogin(c, u)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// set key expiration
	exp := time.Now().Add(Config.Server.SessionExpires).Unix()

	// prepare token for the requested user
	token := token.New(token.SessToken, user.Login)

	// sign the token
	tokenstr, err := token.SignExpires(user.Hash, exp)

	// Enjoy!
	c.JSON(200, gin.H{
		"user":  u,
		"token": tokenstr,
	})
}

// gittoken endpoint return the requested user and it's git api token
func BackstageUserGitToken(c *gin.Context) {
	u := c.Params.ByName("user")

	user, err := store.GetUserLogin(c, u)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.JSON(200, gin.H{
		"user":  u,
		"token": user.Token,
	})
}
