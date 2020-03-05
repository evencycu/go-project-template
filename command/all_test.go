package command

import (
	"testing"
)

func TestServercmd(t *testing.T) {
	initTracer()
	ActivateProfile()
}

func TestAllcmd(t *testing.T) {
	NewAPICmd().Execute()
	NewVersionCmd().Execute()
	Execute()
}
