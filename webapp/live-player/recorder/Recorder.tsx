import {
  Fragment,
  useEffect,
  useRef,
  useState,
  KeyboardEvent
} from "react";
import {
  Box,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  Typography
} from "@mui/material";
import { Cameraswitch, VideocamOff, AddPhotoAlternate, VideoCall } from "@mui/icons-material";
import { takePic, takeVideo, wsMp4fStreamFor } from "webapp/client/livestream";
import { useObservableEffect, useRequest } from "webapp/utils";
import { WsMP4fPlayer } from "webapp/codec/WsMP4fPlayer";
import { StreamStatusProvider, useStreamStatus } from "./SteamStatus";
import { interval, tap } from "rxjs";
import { format } from "date-fns";
import zhCN from "date-fns/locale/zh-CN";

export interface Source {
  id: string;
  name: string;
}

export interface RecorderProps {
  streams: Source[];
}

export interface CameraSwitchProps {
  selectedID: string,
  sources: Source[],
  onSelected: (id: string) => void
}

const ignoreKeyboardKeys = (cb: () => void) => <T extends HTMLElement>(event: KeyboardEvent<T>) => {
  if (
    event.type === "keydown" &&
    (event.key === "Tab" || event.key === "Shift")
  ) {
    return;
  }
  cb();
};

export const CameraSwitch = ({ selectedID, sources, onSelected }: CameraSwitchProps) => {
  const [open, setOpen] = useState(false);
  const toggleDrawer = (open: boolean) => () => {
    setOpen(open);
  };

  return (
    <Fragment>
      <IconButton
        aria-label="切换摄像头"
        onClick={toggleDrawer(true)}
      >
        <Cameraswitch fontSize="inherit" />
      </IconButton>
      <Drawer
        anchor={"left"}
        open={open}
        onClose={toggleDrawer(false)}
      >
        <Box
          sx={{ width: 260 }}
          onClick={toggleDrawer(false)}
          onKeyDown={ignoreKeyboardKeys(toggleDrawer(false))}
          role="presentation"
        >
          <List dense>
            {sources.map((s) => (
              <ListItemButton
                selected={selectedID === s.id}
                key={s.id}
                onClick={(e) => {
                  onSelected(s.id);
                }}
              >
                <ListItemText primary={s.name} />
              </ListItemButton>
            ))}
          </List>
        </Box>
      </Drawer>
    </Fragment>
  );
};

const toWS = (url: string) => "ws:" + url.replaceAll("http:", "").replaceAll("https:", "");

export const Video = ({ url }: { url: string }) => {
  const ref = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (!url) {
      return;
    }

    const player = new WsMP4fPlayer(ref.current!);

    void player.load(toWS(url));

    return () => {
      player.destroy();
    };
  }, [url]);

  return <video ref={ref} width={"100%"} muted={true} />;
};

const RecordingController = ({ id }: { id: string }) => {
  const status = useStreamStatus();

  const takeVideo$ = useRequest(takeVideo);

  const recording = !!status.observers["Video"];

  return <>
    {recording ? (
      <IconButton
        aria-label="结束录制"
        onClick={() => takeVideo$.next({ id, stop: true })}
      >
        <VideocamOff fontSize="inherit" />
      </IconButton>
    ) : (
      <IconButton
        aria-label="开始录制"
        onClick={() => takeVideo$.next({ id })}
      >
        <VideoCall fontSize="inherit" />
      </IconButton>
    )}
  </>;
};

const TakePicController = ({ id }: { id: string }) => {
  const request$ = useRequest(takePic);

  return (
    <IconButton
      aria-label="截图"
      onClick={() => request$.next({ id })}
    >
      <AddPhotoAlternate fontSize="inherit" />
    </IconButton>
  );
};

const CurrentTime = () => {
  const [data, setState] = useState(() => Date.now());

  useObservableEffect(() => {
    return interval(1000).pipe(tap(() => setState(Date.now())));
  }, []);

  return (
    <Box sx={{ paddingLeft: 1, paddingRight: 1 }}>
      <Typography variant="body2">
        {format(data, "yyyy-MM-dd HH:mm:ss", { locale: zhCN })}
      </Typography>
    </Box>
  );
};

export const Recorder = ({ streams }: RecorderProps) => {
  const [id, setId] = useState(streams[0].id);
  const wsMp4fStream$ = useRequest(wsMp4fStreamFor);

  const url = wsMp4fStream$.toHref({ id: id });

  return (
    <Box>
      <Video url={url} />
      <StreamStatusProvider id={id}>
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between"
          }}
        >
          <CameraSwitch sources={streams} selectedID={id} onSelected={setId} />
          <Box sx={{ flex: 1 }}>
            <CurrentTime />
          </Box>
          <RecordingController id={id} />
          <TakePicController id={id} />
        </Box>
      </StreamStatusProvider>
    </Box>
  );
};


