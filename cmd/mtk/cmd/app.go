package cmd

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/media-toolkit/version"
	"log"
	"os"
)

var app = cli.NewApp("mtk", version.Version, &VerboseFlags{})

type VerboseFlags struct {
	V int `flag:"!verbose,v" desc:"verbose level"`
}

func (v *VerboseFlags) PreRun(ctx context.Context) context.Context {
	stdr.SetVerbosity(v.V)
	return logr.NewContext(ctx, NewLogger())
}

func NewLogger() (l logr.Logger) {
	return stdr.New(log.New(os.Stdout, "[mtk] ", log.Flags()))
}

func Run(ctx context.Context) error {
	return cli.Execute(ctx, app, os.Args[1:])
}
