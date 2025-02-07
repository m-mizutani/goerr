package main

import (
	"log/slog"

	"github.com/m-mizutani/goerr/v2"
)

type object struct {
	id    string
	color string
}

func (o *object) Validate() error {
	eb := goerr.NewBuilder().With(goerr.Value("id", o.id))

	if o.color == "" {
		return eb.New("color is empty")
	}

	return nil
}

func main() {
	obj := &object{id: "object-1"}

	if err := obj.Validate(); err != nil {
		slog.Default().Error("Validation error", "err", err)
	}
}
