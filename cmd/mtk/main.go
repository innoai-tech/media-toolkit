package main

import (
	"context"
	"github.com/innoai-tech/media-toolkit/cmd/mtk/cmd"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		panic(err)
	}
}
