import { Fragment, useEffect, useRef, useState, KeyboardEvent } from "react";
import FlvJS from "flv.js";
import { Box, Drawer, IconButton, List, ListItemButton, ListItemText } from "@mui/material";
import { Cameraswitch, VideocamOff, Videocam } from "@mui/icons-material";

export interface Source {
  id: string;
  name: string;
  src: string;
}

export interface RecorderProps {
  sources: Source[];
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

export const Recorder = ({ sources }: RecorderProps) => {
  const ref = useRef<HTMLVideoElement>(null);
  const [recording, updateRecording] = useState(false);
  const [playingID, setPlayingID] = useState(sources[0].id);

  useEffect(() => {
    const src = sources.filter((s) => s.id === playingID)[0]?.src;
    if (!src) {
      return;
    }

    const player = FlvJS.createPlayer({
      type: "flv",
      url: src
    });

    ref.current!.muted = true;

    // player.on(flvjs.Events.METADATA_ARRIVED, (metadata) => console.log(metadata));
    player.on(FlvJS.Events.STATISTICS_INFO, (info) => console.log(info));

    player.attachMediaElement(ref.current!);

    player.load();
    player.play();

    // player.on("STATISTICS_INFO", console.log);

    // ref.current!.addEventListener("click", () => {
    //   ref.current!.muted = false
    // })

    return () => {
      player.detachMediaElement();
    };
  }, [playingID]);

  console.log(playingID);

  return (
    <Box>
      <video ref={ref} width={"100%"} muted={true} />
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between"
        }}
      >
        <CameraSwitch sources={sources} selectedID={playingID} onSelected={setPlayingID} />
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
