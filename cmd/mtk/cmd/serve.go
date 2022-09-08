package cmd

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"

	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/media-toolkit/internal/liveplayer"
)

func init() {
	cli.Add(app, &Serve{})
}

type ServeFlags struct {
	Addr       string `flag:"addr" default:":777" desc:"serve address"`
	ConfigFile string `flag:"config,c" desc:"config file"`
}

type Serve struct {
	cli.Name `desc:"serve"`
	ServeFlags
}

func (p *Serve) Run(ctx context.Context) error {
	streams, err := core.LoadStreams(p.ConfigFile)
	if err != nil {
		return err
	}
	player := &liveplayer.StreamPlayer{
		Addr:    p.Addr,
		Streams: streams,
	}
	return player.Serve(ctx)
}
