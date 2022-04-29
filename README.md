# MediaToolkit

## `streams.json`

```json
[
  {
    "id": "<name>",
    "name": "<name>",
    "rtsp": "rtsp://username:password@192.168.1.66"
  }
]
```

```cue
// Through KubePkg
import (
	"github.com/innoai-tech/media-toolkit/cuepkg/mtk"
	"github.com/innoai-tech/runtime/cuepkg/kube"
)

_mtk: kube.#KubePkg & {
	namespace: "default"
	app:      mtk.#MTK & {
		streams: "x": {
			"name": "<name>"
			"rtsp": "rtsp://username:password@192.168.1.66"
		}
	}
}

_mtk.kube

```