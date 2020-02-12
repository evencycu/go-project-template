package err

import (
	"math/rand"

	"gitlab.com/cake/gopkg"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	causeList = []string{"network", "format", "resource"}

	constCodeErr   = gopkg.NewCarrierCodeError(9996666, "const error")
	getErrFuncList = []func() error{
		randomSessionErr,
		randomNumberErr,
		func() error { return constCodeErr },
	}
)

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randomSessionErr() error {
	return sessionError{
		cause:   causeList[rand.Intn(len(causeList))],
		session: randStringRunes(10),
		err:     sessionErrList[rand.Intn(len(sessionErrList))],
	}
}

func randomNumberErr() error {
	return numberError{
		number: rand.Intn(100),
		err:    numberErrList[rand.Intn(len(numberErrList))],
	}
}
