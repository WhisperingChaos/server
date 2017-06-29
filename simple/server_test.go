package server

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/WhisperingChaos/termChan"
)

func TestUP(t *testing.T) {
	term := termChan.NewAck(termChan.Start())
	debug := log.New(os.Stderr, "simple server:", log.LstdFlags|log.Lshortfile)
	var opts Opts
	opts.Routes = append(opts.Routes, happy(t))
	Start(opts, term, debug)
	term.Wait()
}

func happy(t *testing.T) (rt RouteHandler) {
	fh := func(w http.ResponseWriter, r *http.Request) {
		t.Log(os.Stderr, "Processing happy request.")
		defer r.Body.Close()
		return
	}
	rt.Route = `/happy`
	rt.Handler = fh
	return
}
