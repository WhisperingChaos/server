package oauth

import (
	"log"
	"net/http"

	"github.com/WhisperingChaos/terminator"

	"github.com/RangelReale/osin"
	ex "github.com/RangelReale/osin/example"
	ss "github.com/WhisperingChaos/server/simple"
)

type Opts struct {
	ss.Opts
	TokenExpireInterval int32
}

func Start(opts Opts, term terminator.Isync, debug *log.Logger) {
	term.Add(1)
	dbgInit(debug)
	dbg.Println("oauth server configure handlers start")
	oasvr := oauthServerConfig(opts.TokenExpireInterval)
	// Access token endpoint
	rhToken := ss.RouteHandler{
		Route:   "/token",
		Handler: accessTokenHandlerConfig(oasvr),
	}
	rhAuthorize := ss.RouteHandler{
		Route:   "/authorize",
		Handler: authorizeConfig(oasvr),
	}
	opts.Opts.Routes = append(opts.Opts.Routes, rhAuthorize, rhToken)
	dbg.Println("oauth server configure handlers done")
	dbg.Println("oauth server start generic server framework to handle the requests")
	ss.Start(opts.Opts, term, debug)
}

// private --------------------------------------------------------------------

func oauthServerConfig(tokenExpireInterval int32) (oasvr *osin.Server) {
	// TestStorage implements the "osin.Storage" interface
	oasvr = osin.NewServer(osin.NewServerConfig(), ex.NewTestStorage())
	oasvr.Config.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.CLIENT_CREDENTIALS}
	oasvr.Config.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	if tokenExpireInterval == 0 {
		tokenExpireInterval = 3600 // 1 hour
	}
	oasvr.Config.AccessExpiration = tokenExpireInterval
	return
}

// Authorization code endpoint
func authorizeConfig(oasvr *osin.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dbg.Println("Authorization start")
		defer dbg.Println("Authorization end")
		resp := oasvr.NewResponse()
		defer resp.Close()
		if ar := oasvr.HandleAuthorizeRequest(resp, r); ar != nil {
			// HANDLE LOGIN PAGE HERE
			dbg.Println("Authorization ok")
			ar.Authorized = true
			oasvr.FinishAuthorizeRequest(resp, r, ar)
		}
		osin.OutputJSON(resp, w, r)
	}
}

// Obtain new Access Token
func accessTokenHandlerConfig(oasvr *osin.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dbg.Println("Token start")
		defer dbg.Println("Token end")
		resp := oasvr.NewResponse()
		defer resp.Close()
		dbg.Println("Token obtain")
		if ar := oasvr.HandleAccessRequest(resp, r); ar != nil {
			ar.Authorized = true
			oasvr.FinishAccessRequest(resp, r, ar)
			dbg.Println("Token generated")
			osin.OutputJSON(resp, w, r)
		}
	}
}

var dbg *log.Logger = debugNull()

type nullog struct{}

func dbgInit(debug *log.Logger) {
	if debug != nil {
		dbg = debug
	}
}
func (*nullog) Write(p []byte) (int, error) {
	return len(p), nil
}
func debugNull() *log.Logger {
	return log.New(new(nullog), "", 0)
}
