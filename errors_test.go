package goerr_test

import (
	"fmt"
	"testing"

	"github.com/m-mizutani/goerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func oops() *goerr.Error {
	return goerr.New("omg")
}

func normalError() error {
	return fmt.Errorf("red")
}

func wrapError() *goerr.Error {
	err := normalError()
	return goerr.Wrap(err, "orange")
}

func TestNew(t *testing.T) {
	err := oops()
	assert.Contains(t, fmt.Sprintf("%+v", err), "goerr_test.oops")
	assert.Contains(t, err.Error(), "omg")
}

func TestWrapError(t *testing.T) {
	err := wrapError()
	st := fmt.Sprintf("%+v", err)
	assert.Contains(t, st, "github.com/m-mizutani/goerr_test.wrapError\n")
	assert.Contains(t, st, "github.com/m-mizutani/goerr_test.TestWrapError\n")
	assert.NotContains(t, st, "github.com/m-mizutani/goerr_test.normalError\n")
	assert.Contains(t, err.Error(), "orange: red")
}

func TestStackTrace(t *testing.T) {
	err := oops()
	st := err.Stacks()
	require.Equal(t, 4, len(st))
	assert.Equal(t, "github.com/m-mizutani/goerr_test.oops", st[0].Func)
	assert.Regexp(t, `/goerr/errors_test\.go$`, st[0].File)
	assert.Equal(t, 13, st[0].Line)
}

func TestMultileWrap(t *testing.T) {
	err1 := oops()
	err2 := goerr.Wrap(err1)
	assert.Equal(t, err1, err2)

	err3 := goerr.Wrap(err1, "some message")
	assert.NotEqual(t, err1, err3)
}
