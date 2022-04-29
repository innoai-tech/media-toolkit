package livestream

import (
	"context"
	"github.com/deepch/vdk/av"
	"io"
)

type StreamWriterSwitch interface {
	Enable(ctx context.Context)
	Disable(ctx context.Context)
}

type StreamWriter interface {
	Enabled() bool
	SetCodecData(data []av.CodecData)
	WritePacket(pkt *Packet)
	io.Closer
}

func MultiStreamWriter(writers ...StreamWriter) StreamWriter {
	return &multiStreamWriter{writers: writers}
}

type multiStreamWriter struct {
	Switch
	writers []StreamWriter
}

func (dw *multiStreamWriter) Enabled() bool {
	enabled := false
	for i := range dw.writers {
		if dw.writers[i].Enabled() {
			enabled = true
		}
	}
	return enabled
}

func (dw *multiStreamWriter) SetCodecData(data []av.CodecData) {
	for i := range dw.writers {
		dw.writers[i].SetCodecData(data)
	}
}

func (dw *multiStreamWriter) WritePacket(pkt *Packet) {
	for i := range dw.writers {
		dw.writers[i].WritePacket(pkt)
	}
}

func (dw *multiStreamWriter) Close() error {
	for i := range dw.writers {
		_ = dw.writers[i].Close()
	}
	return nil
}
