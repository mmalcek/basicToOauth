package main

import (
	"context"
	"encoding/base64"
	"errors"
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

// basicAuthChallenge is advertised to clients on a 401 so legacy Basic-auth
// clients know to (re)send credentials, which the proxy then exchanges for an
// OAuth token.
const basicAuthChallenge = `Basic realm="basicToOauth"`

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
	proxy.ModifyResponse = func(resp *http.Response) error {
		injectBasicChallenge(resp, c.Request.Header.Get("Authorization"))
		return nil
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// injectBasicChallenge re-advertises a Basic challenge so legacy clients know
// to send credentials — but ONLY for requests that arrived without auth.
// Re-challenging a request that already carried (bad) credentials loops the
// client and hammers Azure AD with repeated bad-cred token requests
// (smart-lockout risk on a live mailbox).
func injectBasicChallenge(resp *http.Response, clientAuth string) {
	if resp.StatusCode == http.StatusUnauthorized && clientAuth == "" {
		resp.Header.Set("WWW-Authenticate", basicAuthChallenge)
	}
}

// acquireToken is the seam used to obtain an Azure token. In production it is
// the real getAzureToken; tests override it to avoid live Azure AD calls.
var acquireToken = getAzureToken

func getAuthHeader(authHeader string) string {
	if strings.Split(authHeader, " ")[0] != "Basic" { // If anythig else than Basic auth, return original header
		return authHeader
	}
	currHeader := strings.Split(authHeader, " ")
	if currHeader[0] == "Basic" && len(currHeader) == 2 { // If authHeader seems to be Basic auth
		mapToken := tokensMap.get(currHeader[1])
		if mapToken != nil { // If token is in map
			if mapToken.expire.Unix()-60 < time.Now().Unix() { // If token is about to expire try to get new one
				newToken, err := acquireToken(currHeader[1])
				if err != nil {
					logger.Warningf("ERR-getAzureToken: %v", err.Error())
					return authHeader
				}
				tokensMap.add(currHeader[1], newToken) // Add new token to map
			}
			return "Bearer " + tokensMap.get(currHeader[1]).token // Return token from map
		} else { // If token is not in map
			newToken, err := acquireToken(currHeader[1]) // Get new token
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
		return nil, errors.New("basicAuthParseFailed")
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
