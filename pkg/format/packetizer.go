package format

import (
	"encoding/binary"
	"time"
)

type AVPacket struct {
	Idx        int
	Time       time.Duration
	IsKeyFrame bool
	Data       []byte

	SPS []byte
	PPS []byte
	VPS []byte // H265 only
}

type Packetizer struct {
	Time time.Duration
}

func (p *Packetizer) Packetize(payload []byte, samples uint32) *AVPacket {
	p.Time += time.Duration(samples) * time.Microsecond

	pkt := &AVPacket{}
	pkt.Data = make([]byte, 0, len(payload))
	pkt.Time = p.Time

	emitNalus(payload, func(nalu []byte) {
		naluType := nalu[0] & 0x1f

		switch {
		case naluType >= 1 && naluType <= 5:
			pkt.IsKeyFrame = naluType == 5
			pkt.Data = append(pkt.Data, append(binSize(len(nalu)), nalu...)...)
		case naluType == 7: // sps
			pkt.SPS = nalu
		case naluType == 8: // pps
			pkt.PPS = nalu
		case naluType == 6: // sei
			// skip
		}
	})

	return pkt
}

func emitNalus(nals []byte, emit func([]byte)) {
	nextInd := func(nalu []byte, start int) (indStart int, indLen int) {
		zeroCount := 0

		for i, b := range nalu[start:] {
			if b == 0 {
				zeroCount++
				continue
			} else if b == 1 {
				if zeroCount >= 2 {
					return start + i - zeroCount, zeroCount + 1
				}
			}
			zeroCount = 0
		}
		return -1, -1
	}

	nextIndStart, nextIndLen := nextInd(nals, 0)
	if nextIndStart == -1 {
		emit(nals)
	} else {
		for nextIndStart != -1 {
			prevStart := nextIndStart + nextIndLen
			nextIndStart, nextIndLen = nextInd(nals, prevStart)
			if nextIndStart != -1 {
				emit(nals[prevStart:nextIndStart])
			} else {
				// Emit until end of stream, no end indicator found
				emit(nals[prevStart:])
			}
		}
	}
}

func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
