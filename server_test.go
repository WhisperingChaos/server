package server

import (
	"testing"

	"github.com/WhisperingChaos/termChan"
)

func TestUP(t *testing.T) {
	term := termChan.NewAck(termChan.Start())
	Start(TarServer, term)
	term.Wait()
}
