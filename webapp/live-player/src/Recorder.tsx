import { Fragment, useEffect, useRef, useState, KeyboardEvent, useMemo } from "react";
import { Box, Drawer, IconButton, List, ListItemButton, ListItemText } from "@mui/material";
import { Cameraswitch, VideocamOff, Videocam } from "@mui/icons-material";
import { toHref } from "../client/client";
import { wsMp4fStreamFor } from "../client/livestream";
import { MsePlayer } from "../codec/MsePlayer";

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

    const player = new MsePlayer(ref.current!);

    void player.load(toWS(url));

    return () => {
      player.destroy();
    };
  }, [url]);

  return <video ref={ref} width={"100%"} muted={true} />;
};

export const Recorder = ({ streams }: RecorderProps) => {
  const [recording, updateRecording] = useState(false);
  const [playingID, setPlayingID] = useState(streams[0].id);

  const url = toHref(wsMp4fStreamFor({ id: playingID }));

  return (
    <Box>
      <Video url={url} />
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between"
        }}
      >
        <CameraSwitch sources={streams} selectedID={playingID} onSelected={setPlayingID} />
        {recording ? (
          <IconButton
            aria-label="结束录制"
            onClick={() => updateRecording(false)}
          >
            <VideocamOff fontSize="inherit" />
          </IconButton>
        ) : (
          <IconButton
            aria-label="开始录制"
            onClick={() => updateRecording(true)}
          >
            <Videocam fontSize="inherit" />
          </IconButton>
        )}
      </Box>
    </Box>
  );
};
