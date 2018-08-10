package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/drone/drone/model"
	"github.com/drone/drone/shared/token"
	"github.com/drone/drone/store"
	"github.com/gin-gonic/gin"
)

func whitelistedRemoteAddr(c *gin.Context) bool {
	remoteip := strings.Split(c.Request.RemoteAddr, ":")[0]

	if remoteip != Config.Backstage.WhitelistedRemoteAddr {
		return false
	}

	return true
}

// apitoken endpoint returns the requested user and a token signed by user.Hash
func BackstageUserApiToken(c *gin.Context) {

	if !whitelistedRemoteAddr(c) {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("Unprivileged access from %s", c.Request.RemoteAddr))
		return
	}

	l := c.Params.ByName("login")

	user, err := store.GetUserLogin(c, l)

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
		"login": l,
		"token": tokenstr,
	})
}

// gittoken endpoint return the requested user and it's git api token
func BackstageUserGitToken(c *gin.Context) {

	if !whitelistedRemoteAddr(c) {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("Unprivileged access from %s", c.Request.RemoteAddr))
		return
	}

	l := c.Params.ByName("login")

	user, err := store.GetUserLogin(c, l)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.JSON(200, gin.H{
		"login": l,
		"token": user.Token,
	})
}

// hooktoken endpoint return token for requested owner/repo
func BackstageGetRepoHook(c *gin.Context) {

	if !whitelistedRemoteAddr(c) {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("Unprivileged access from %s", c.Request.RemoteAddr))
		return
	}

	// grab owner param
	ownerstr := c.Params.ByName("owner")

	// grab name param
	namestr := c.Params.ByName("name")

	fullname := ownerstr + "/" + namestr

	t := token.New(token.HookToken, fullname)

	repo, err := store.GetRepoOwnerName(c, ownerstr, namestr)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"full_name": fullname,
		"token":     sig,
	})
}

// gittoken endpoint return the requested user and it's git api token
func BackstageGetRepoConfig(c *gin.Context) {

	if !whitelistedRemoteAddr(c) {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("Unprivileged access from %s", c.Request.RemoteAddr))
		return
	}

	// grab owner param
	ownerstr := c.Params.ByName("owner")

	// grab name param
	namestr := c.Params.ByName("name")

	// get repo object so that we can locate related configuration
	repo, err := store.GetRepoOwnerName(c, ownerstr, namestr)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// get current configuration object
	currentconfig, err := Config.Storage.Config.ConfigFindFirst(repo)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.JSON(200, *currentconfig)
}

// gittoken endpoint return the requested user and it's git api token
func BackstagePostRepoConfig(c *gin.Context) {

	if !whitelistedRemoteAddr(c) {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("Unprivileged access from %s", c.Request.RemoteAddr))
		return
	}

	// grab owner param
	ownerstr := c.Params.ByName("owner")

	// grab name param
	namestr := c.Params.ByName("name")

	// get repo object so that we can locate related configuration
	repo, err := store.GetRepoOwnerName(c, ownerstr, namestr)

	// get current configuration object
	currentconfig, err := Config.Storage.Config.ConfigFindFirst(repo)

	var reader io.Reader = c.Request.Body

	raw, err := ioutil.ReadAll(reader)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// do not store config for repository we know nothing about
	if err != nil {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("No such repository as %s/%s", ownerstr, namestr))
		return
	}

	conf := &model.Config{
		RepoID: repo.ID,
		Data:   string(raw),
		Hash:   shasum(raw),
	}

	// update conf object with the found configuration so we can properly execute ConfigUpdate
	if currentconfig.ID > 0 {
		conf.ID = currentconfig.ID
		err = Config.Storage.Config.ConfigUpdate(conf)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(200, *conf)
		return
	}

	// looks like we need to create a new one. Welcome!
	err = Config.Storage.Config.ConfigCreate(conf)

	c.JSON(200, *conf)
}

// /stats/builds/failed endpoint return total number of system wide failed builds
func BackstageStatsFailedBuildsCount(c *gin.Context) {
	count, err := store.GetBuildFailedCount(c)

	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"value": count,
	})
}

// /stats/builds/succeeded endpoint return total number of system wide succeeded builds
func BackstageStatsSucceededBuildsCount(c *gin.Context) {
	count, err := store.GetBuildSucceededCount(c)

	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"value": count,
	})
}
