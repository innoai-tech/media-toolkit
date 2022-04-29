export class MP4fSourceBuffer {
  static create = (
    mediaSource: MediaSource,
    codecs: string
  ): MP4fSourceBuffer => {
    const mimeCodec = `video/mp4; codecs="${codecs}"`;

    if (!MediaSource.isTypeSupported(mimeCodec)) {
      throw `unsupported ${mimeCodec}`;
    }

    return new MP4fSourceBuffer(
      mediaSource,
      mediaSource.addSourceBuffer(mimeCodec)
    );
  };

  private streamingStarted = false;
  private buffer = [] as ArrayBuffer[];

  constructor(
    private mediaSource: MediaSource,
    private sourceBuffer: SourceBuffer
  ) {
    this.sourceBuffer.mode = "segments";
    this.sourceBuffer.addEventListener("updateend", this.load);
  }

  destroy = () => {
    try {
      this.mediaSource.removeSourceBuffer(this.sourceBuffer);
    } catch (_) {}
    this.sourceBuffer.removeEventListener("updateend", this.load);

    this.buffer = [];
    this.streamingStarted = true;
  };

  writePacket = (data: Uint8Array) => {
    if (!this.streamingStarted) {
      this.appendSourceBuffer(data);
      this.streamingStarted = true;
      return;
    }
    this.buffer.push(data);
    this.load();
  };

  private appendSourceBuffer = (buf: Uint8Array) => {
    try {
      this.sourceBuffer.appendBuffer(buf);
    } catch (err) {
      console.error(err);
    }
  };

  private load = () => {
    if (!this.sourceBuffer.updating) {
      if (this.buffer.length > 0) {
        this.appendSourceBuffer(new Uint8Array(this.buffer.shift()!));
      } else {
        this.streamingStarted = false;
      }
    }
  };
}

export class WsMP4fPlayer {
  private ms = new MediaSource();

  constructor(public $video: HTMLVideoElement) {}

  url: string = "";

  load(url: string) {
    this.url = url;

    this.effect(() => {
      this.ms.addEventListener("sourceopen", this.start);
      return () => this.ms.removeEventListener("sourceopen", this.start);
    });

    this.effect(() => {
      const onError = () => {
        console.log(
          "Error " +
            this.$video.error!.code +
            "; details: " +
            this.$video.error!.message
        );
      };

      this.$video.addEventListener("error", onError);
      return () => {
        this.$video.removeEventListener("error", onError);
      };
    });

    this.$video.src = window.URL.createObjectURL(this.ms);
    this.$video.pause();
    return this.$video.play();
  }

  private destroyed = false;
  private cleanups = [] as Array<() => void>;

  private effect(fn: () => () => void) {
    const cleanup = fn();
    this.cleanups.push(cleanup);
  }

  destroy() {
    this.destroyed = true;
    for (const cleanup of this.cleanups) {
      cleanup();
    }
    this.cleanups = [];
  }

  private start = () => {
    this.effect(() => {
      const ws = new WebSocket(this.url);

      ws.binaryType = "arraybuffer";
      ws.onopen = () => {
        console.log(`${this.url} connect`);
      };

      let buf: MP4fSourceBuffer | null;

      ws.onmessage = (event) => {
        const data = new Uint8Array(event.data);

        if (data[0] === 9) {
          buf = MP4fSourceBuffer.create(
            this.ms,
            new TextDecoder("utf-8").decode(data.slice(1))
          );
        } else if (data[0] === 8) {
          const evt = new CustomEvent("METADATA", {
            detail: JSON.parse(new TextDecoder("utf-8").decode(data.slice(1))),
          });
          this.$video && this.$video.dispatchEvent(evt);
        } else {
          buf?.writePacket(data);
        }

        if (document.hidden && this.$video.buffered.length) {
          this.$video.currentTime =
            this.$video.buffered.end(this.$video.buffered.length - 1) - 1;
        }
      };

      ws.onclose = () => {
        buf?.destroy();
        buf = null;
        console.log(`${this.url} closed`);

        if (!this.destroyed) {
          this.effect(() => {
            console.log("re connect 3 seconds later");
            const timer = setTimeout(() => {
              void this.load(this.url);
            }, 3_000);
            return () => clearTimeout(timer);
          });
        }
      };

      return () => {
        ws.close();
      };
    });
  };
}
