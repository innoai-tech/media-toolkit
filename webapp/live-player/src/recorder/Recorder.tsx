import {Fragment, KeyboardEvent, useEffect, useRef, useState} from "react";
import {
    Box,
    Drawer,
    IconButton,
    List,
    ListItemButton,
    ListItemText,
    Typography,
} from "@mui/material";
import {
    AddPhotoAlternate,
    Cameraswitch,
    VideoCall,
    VideocamOff,
} from "@mui/icons-material";
import {takePic, takeVideo, wsMp4fStreamFor} from "../client/livestream";
import {useObservable, useObservableEffect, useRequest} from "../utils";
import {WsMP4fPlayer} from "../codec/WsMP4fPlayer";
import {StreamStatusProvider, useStreamStatus} from "./SteamStatus";
import {fromEvent, interval, tap} from "rxjs";
import {format, parseISO} from "date-fns";
import zhCN from "date-fns/locale/zh-CN";
import {omit} from "@innoai-tech/lodash";
import {useStateSubject} from "@innoai-tech/reactutil";

export interface Source {
    id: string;
    name: string;
}

export interface RecorderProps {
    streams: Source[];
}

export interface CameraSwitchProps {
    selectedID: string;
    sources: Source[];
    onSelected: (id: string) => void;
}

const ignoreKeyboardKeys =
    (cb: () => void) => <T extends HTMLElement>(event: KeyboardEvent<T>) => {
        if (
            event.type === "keydown" &&
            (event.key === "Tab" || event.key === "Shift")
        ) {
            return;
        }
        cb();
    };

export const CameraSwitch = ({
                                 selectedID,
                                 sources,
                                 onSelected,
                             }: CameraSwitchProps) => {
    const [open, setOpen] = useState(false);
    const toggleDrawer = (open: boolean) => () => {
        setOpen(open);
    };

    return (
        <Fragment>
            <IconButton aria-label="切换摄像头" onClick={toggleDrawer(true)}>
                <Cameraswitch fontSize="inherit"/>
            </IconButton>
            <Drawer anchor={"left"} open={open} onClose={toggleDrawer(false)}>
                <Box
                    sx={{width: 260}}
                    onClick={toggleDrawer(false)}
                    onKeyDown={ignoreKeyboardKeys(toggleDrawer(false))}
                    role="presentation"
                >
                    <List dense={true}>
                        {sources.map(
                            (s) => (
                                <ListItemButton
                                    selected={selectedID === s.id}
                                    key={s.id}
                                    onClick={() => {
                                        onSelected(s.id);
                                    }}
                                >
                                    <ListItemText primary={s.name}/>
                                </ListItemButton>
                            ),
                        )}
                    </List>
                </Box>
            </Drawer>
        </Fragment>
    );
};

const toWS = (url: string) => {
    return url.replaceAll("http:", "ws:").replaceAll("https:", "wss:");
};

const RecordingController = ({id}: { id: string }) => {
    const status$ = useStreamStatus();

    const takeVideo$ = useRequest(takeVideo);

    useObservableEffect(() => {
        return takeVideo$.pipe(
            tap((resp) => {
                return status$.next(
                    (status) => ({
                        ...status,
                        observers: resp.config.inputs.stop
                            ? omit(status.observers, "Video")
                            : {
                                ...status.observers,
                                Video: 1,
                            },
                    }),
                );
            }),
        );
    }, []);

    const status = useObservable(status$);

    const recording = !!status?.observers["Video"];

    return (
        <>
            {recording ? (
                <IconButton
                    aria-label="结束录制"
                    onClick={() => takeVideo$.next({id, stop: true})}
                >
                    <VideocamOff fontSize="inherit"/>
                </IconButton>
            ) : (
                <IconButton aria-label="开始录制" onClick={() => takeVideo$.next({id})}>
                    <VideoCall fontSize="inherit"/>
                </IconButton>
            )}
        </>
    );
};

const TakePicController = ({id}: { id: string }) => {
    const request$ = useRequest(takePic);

    return (
        <IconButton aria-label="截图" onClick={() => request$.next({id})}>
            <AddPhotoAlternate fontSize="inherit"/>
        </IconButton>
    );
};

const formatDateTime = (date: Date | number) =>
    format(date, "yyyy-MM-dd HH:mm:ss", {locale: zhCN});

const CurrentTimeFromMetadata = () => {
    const status$ = useStreamStatus();
    const status = useObservable(status$);

    return (
        <Box sx={{paddingLeft: 1, paddingRight: 1}}>
            <Typography variant="body2">
                {status?.time && formatDateTime(status.time)}
            </Typography>
        </Box>
    );
};


const CurrentTime = () => {
    const now$ = useStateSubject(() => new Date())

    useObservableEffect(() => {
        return interval(1000).pipe(tap(() => {
            now$.next(new Date())
        }))
    });

    const now = useObservable(now$);

    return (
        <Box sx={{paddingLeft: 1, paddingRight: 1}}>
            <Typography variant="body2" sx={{fontSize: 10}}>
                {formatDateTime(now)}
            </Typography>
        </Box>
    );
};


export const Video = ({url}: { url: string }) => {
    const status$ = useStreamStatus();

    const ref = useRef<HTMLVideoElement>(null);

    useObservableEffect(() => {
        if (!ref.current) {
            return;
        }
        return fromEvent(ref.current, "METADATA").pipe(
            tap((evt) => {
                const metadata = (evt as CustomEvent).detail;

                status$.next(
                    (status) => ({
                        ...status,
                        time: parseISO(metadata.at),
                        observers: metadata.observers,
                    }),
                );
            }),
        );
    }, []);

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

    return <video ref={ref} width={"100%"} muted={true}/>;
};

export const Recorder = ({streams}: RecorderProps) => {
    const [id, setId] = useState(streams[0]!.id);
    const wsMp4fStream$ = useRequest(wsMp4fStreamFor);

    const url = wsMp4fStream$.toHref({id: id});

    return (
        <Box>
            <StreamStatusProvider id={id}>
                <Video url={url}/>
                <Box
                    sx={{
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "space-between",
                    }}
                >
                    <CameraSwitch sources={streams} selectedID={id} onSelected={setId}/>
                    <Box sx={{flex: 1}}>
                        <CurrentTimeFromMetadata/>
                        <CurrentTime/>
                    </Box>
                    <RecordingController id={id}/>
                    <TakePicController id={id}/>
                </Box>
            </StreamStatusProvider>
        </Box>
    );
};
