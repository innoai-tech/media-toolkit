package cmd

import (
	"context"

	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/media-toolkit/pkg/live"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
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
	streams, err := livestream.LoadStreams(p.ConfigFile)
	if err != nil {
		return err
	}
	player := &live.StreamPlayer{
		Addr:    p.Addr,
		Streams: streams,
	}
	return player.Serve(ctx)
}
