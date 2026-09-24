package tests

import (
	"github.com/goravel/framework/testing"

	"dev.jevido/jevidocs/services/api/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
