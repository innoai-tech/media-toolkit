package cmd

import "github.com/innoai-tech/infra/pkg/cli"

var cmdLive = cli.Add(app, &Live{})

type Live struct {
	cli.Name `desc:"live stream"`
}
