package main

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/gin-gonic/gin"
)

func (p *program) run() {
	if err := loadConfig(); err != nil {
		logger.Errorf("ERR-loadConfig: %v", err.Error())
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Any("/*proxyPath", proxy)
	go r.Run(net.JoinHostPort(config.Host, config.Port))
	logger.Infof("basicToOauth version: %s, started on: %s\r\n", VERSION, net.JoinHostPort(config.Host, config.Port))

	go func() { // Check and delete expired tokens every 5 minutes
		for {
			tokensMap.delExpired()
			time.Sleep(10 * time.Minute)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	<-p.exit
}

func proxy(c *gin.Context) {
	remote, err := url.Parse(config.ProxyURL)
	if err != nil {
		logger.Errorf("ERR-remoteURL: %v", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header.Set("Authorization", getAuthHeader(c.Request.Header.Get("Authorization")))
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("proxyPath")
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func getAuthHeader(authHeader string) string {
	if strings.Split(authHeader, " ")[0] != "Basic" { // If anythig else than Basic auth, return original header
		return authHeader
	}
	currHeader := strings.Split(authHeader, " ")
	if currHeader[0] == "Basic" && len(currHeader) == 2 { // If authHeader seems to be Basic auth
		mapToken := tokensMap.get(currHeader[1])
		if mapToken != nil { // If token is in map
			if mapToken.expire.Unix()-60 < time.Now().Unix() { // If token is about to expire try to get new one
				newToken, err := getAzureToken(currHeader[1])
				if err != nil {
					logger.Warningf("ERR-getAzureToken: %v", err.Error())
					return authHeader
				}
				tokensMap.add(currHeader[1], newToken) // Add new token to map
			}
			return "Bearer " + tokensMap.get(currHeader[1]).token // Return token from map
		} else { // If token is not in map
			newToken, err := getAzureToken(currHeader[1]) // Get new token
			if err != nil {
				logger.Warningf("ERR-getAzureToken: %v", err.Error())
				return authHeader
			}
			tokensMap.add(currHeader[1], newToken) // Add new token to map
			return "Bearer " + newToken.token      // Return new token
		}
	}
	return authHeader
}

// Request Azure token
func getAzureToken(baseKey string) (*tToken, error) {
	baseDecode, err := base64.StdEncoding.DecodeString(baseKey)
	if err != nil {
		return nil, err
	}
	baseSplit := strings.Split(string(baseDecode), ":")
	if len(baseSplit) != 2 {
		return nil, err
	}
	publicClientApp, err := public.New(config.ClientID, public.WithAuthority(config.AuthorityURL+config.TenantID))
	if err != nil {
		return nil, err
	}
	result, err := publicClientApp.AcquireTokenByUsernamePassword(context.Background(), config.Scopes, baseSplit[0], baseSplit[1])
	if err != nil {
		return nil, err
	}
	return &tToken{token: result.AccessToken, expire: result.ExpiresOn}, nil
}
