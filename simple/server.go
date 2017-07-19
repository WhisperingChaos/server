package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/WhisperingChaos/termChan"

	"github.com/WhisperingChaos/terminator"
)

type RouteHandler struct {
	Route   string
	Handler func(http.ResponseWriter, *http.Request)
}
type Opts struct {
	RootURL    string // defaults to localhost
	Port       uint16 // defaults to 10000
	Routes     []RouteHandler
	TLSenabled bool               // defaults to don't start TLS server
	PortTLS    uint16             // defaults to 10443
	ClientAuth tls.ClientAuthType // can be NoClientCert, RequestClientCert, or tls.RequireAnyClientCert.  defaults to NoClientCert
	CertPath   string
	KeyPath    string
}

func Start(opts Opts, term terminator.Isync, debug *log.Logger) {
	term.Add(1)
	//	dbgInit(debug)
	start := termChan.New()
	go run(opts, start, term)
	<-start.Chan()
}
func run(opts Opts, start terminator.I, term terminator.Isync) {
	defer term.Done()
	optCheck(&opts)
	serverStart, serverShutdown := manageNewServers(opts)
	serverStart()
	close(start.Chan())
	for term.IsNot() {
		select {
		case <-term.Chan():
			serverShutdown()
		}
	}
}
func optCheck(opts *Opts) {
	if len(opts.Routes) == 0 {
		dbg.Fatalln("no routes/handlers specified for server")
	}
	if opts.RootURL == "" {
		opts.RootURL = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 10000
	}
	if !opts.TLSenabled {
		return
	}
	if opts.PortTLS == 0 {
		opts.PortTLS = 10443
	}
	if opts.CertPath == "" {
		dbg.Fatalln("CertPath not specified.")
	}
	if opts.KeyPath == "" {
		dbg.Fatalln("KeyPath not specified.")
	}
}
func manageNewServers(opts Opts) (start func(), shutdown func()) {
	var serverTLS *http.Server
	var serverHttp *http.Server
	var lck sync.Mutex
	start = func() {
		mux := http.NewServeMux()
		for _, rth := range opts.Routes {
			dbg.Println("Creating handler for route: " + rth.Route)
			mux.HandleFunc(rth.Route, rth.Handler)
		}
		startHTTP := termChan.New()
		go func() {
			<-startHTTP.Chan()
			close(startHTTP.Chan())
			var url = opts.RootURL + ":" + strconv.Itoa(int(opts.Port))
			lserverHttp := &http.Server{Handler: mux, Addr: url}
			lck.Lock()
			serverHttp = lserverHttp
			lck.Unlock()
			dbg.Println("Started HTTP listner on:" + url)
			lserverHttp.ListenAndServe()
		}()
		startHTTP.Chan() <- true
		<-startHTTP.Chan()
		if !opts.TLSenabled {
			return
		}
		startTLS := termChan.New()
		go func() {
			<-startTLS.Chan()
			close(startTLS.Chan())
			var urlTLS = opts.RootURL + ":" + strconv.Itoa(int(opts.PortTLS))
			lserverTLS := &http.Server{
				TLSConfig: &tls.Config{
					ClientAuth: opts.ClientAuth,
				},
				Addr:    urlTLS,
				Handler: mux,
			}
			lck.Lock()
			serverTLS = lserverTLS
			lck.Unlock()
			dbg.Println("Started HTTPS listner on: " + urlTLS)
			lserverTLS.ListenAndServeTLS(opts.CertPath, opts.KeyPath)
		}()
		startTLS.Chan() <- true
		<-startTLS.Chan()
	}
	shutdown = func() {
		lck.Lock()
		defer lck.Unlock()
		if serverHttp != nil {
			serverHttp.Shutdown(nil)
		}
		if serverTLS != nil {
			serverTLS.Shutdown(nil)
		}
	}
	return
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
