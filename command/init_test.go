package command

import (
	"log"
	"os"
	"testing"

	"gitlab.com/cake/goctx"

	dockertest "github.com/ory/dockertest/v3"
)

var (
	dockerPool *dockertest.Pool
)

var (
	testCtx goctx.Context
)

func BeforeTest() {
	testCtx = goctx.Background()
}

func TestMain(m *testing.M) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)
	var p *int
	retCode := 0
	p = &retCode
	BeforeTest()
	defer AfterTest(p)
	*p = m.Run()
}

func AfterTest(ret *int) {
	os.Exit(*ret)
}
