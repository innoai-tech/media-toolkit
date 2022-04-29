package livestream

import (
	"io"

	"github.com/deepch/vdk/av"
)

type StreamObserver interface {
	io.Closer
	SetCodecData(data []av.CodecData)
	WritePacket(pkt Packet)
}
