package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/drone/drone/model"
	"github.com/drone/drone/shared/token"
	"github.com/drone/drone/store"
	"github.com/gin-gonic/gin"
)

const (
	repoConfigField = "data"
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

// gittoken endpoint return the requested user and it's git api token
func BackstageRepoConfig(c *gin.Context) {

	// grab owner param
	ownerstr := c.Params.ByName("owner")

	// grab name param
	namestr := c.Params.ByName("name")

	var reader io.Reader = c.Request.Body

	raw, err := ioutil.ReadAll(reader)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	repoconfig := new(model.BackstageRepoConfig)

	err = json.Unmarshal(raw, repoconfig)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	repo, err := store.GetRepoOwnerName(c, ownerstr, namestr)

	// do not store config for repository we know nothing about
	if err != nil {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("No such repository as %s/%s", ownerstr, namestr))
		return
	}

	confb, _ := base64.StdEncoding.DecodeString(repoconfig.Data)

	conf := &model.Config{
		RepoID: repo.ID,
		Data:   string(confb),
		Hash:   shasum(confb),
	}

	// get current configuration object
	currentconfig, err := Config.Storage.Config.ConfigFindFirst(repo)

	// update conf object with the found configuration so we can properly execute ConfigUpdate
	if err == nil {
		conf.ID = currentconfig.ID
		err = Config.Storage.Config.ConfigUpdate(conf)
		c.JSON(200, *conf)
		return
	}

	// looks like we need to create a new one. Welcome!
	err = Config.Storage.Config.ConfigCreate(conf)

	c.JSON(200, *conf)
}
