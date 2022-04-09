package goerr_test

import (
	"errors"
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
	assert.Equal(t, 14, st[0].Line)
}

func TestMultileWrap(t *testing.T) {
	err1 := oops()
	err2 := goerr.Wrap(err1)
	assert.NotEqual(t, err1, err2)

	err3 := goerr.Wrap(err1, "some message")
	assert.NotEqual(t, err1, err3)
}

func TestErrorCode(t *testing.T) {
	rootErr := goerr.New("something bad")
	baseErr1 := goerr.New("oops").Code("code1")
	baseErr2 := goerr.New("oops").Code("code2")

	newErr := baseErr1.Wrap(rootErr).With("v", 1)

	assert.True(t, errors.Is(newErr, baseErr1))
	assert.NotEqual(t, newErr, baseErr1)
	assert.NotNil(t, newErr.Values()["v"])
	assert.Nil(t, baseErr1.Values()["v"])

	assert.False(t, errors.Is(newErr, baseErr2))
}

func TestPrintable(t *testing.T) {
	cause := errors.New("test")
	err := goerr.Wrap(cause, "oops").Code("E001").With("blue", "five")

	p := err.Printable()
	assert.Equal(t, "oops", p.Message)
	assert.Equal(t, "E001", p.Code)
	assert.Equal(t, cause, p.Cause)
	assert.Equal(t, "five", p.Values["blue"])
}
