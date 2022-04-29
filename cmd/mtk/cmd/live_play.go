package cmd

import (
	"context"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/media-toolkit/pkg/live"
)

func init() {
	cli.Add(cmdLive, &Play{})
}

type Play struct {
	cli.Name `args:"LIVE_STREAM_URI" desc:"live stream address"`
}

func (p *Play) Run(ctx context.Context) error {
	player := &live.Player{}
	return player.Serve(ctx)
}
