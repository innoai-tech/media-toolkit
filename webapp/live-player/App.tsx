import { Box, Container, createTheme, CssBaseline, Paper, ThemeProvider, useMediaQuery } from "@mui/material";
import { Recorder } from "webapp/live-player/recorder";
import { BlobInfoQuerier } from "webapp/live-player/blob";
import { listStreams } from "webapp/client/livestream";
import { useObservable, useRequest } from "webapp/utils";
import { useEffect } from "react";

const theme = createTheme();

export const App = () => {
  const request$ = useRequest(listStreams);
  const resp = useObservable(request$);

  useEffect(() => {
    request$.next(undefined);
  }, []);

  const maxWidthMatched = useMediaQuery("(max-width:600px)");

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box
        component="main"
        sx={{
          display: "flex",
          flexDirection: maxWidthMatched ? "column" : "row",
          overflow: "hidden",
          position: "absolute",
          top: 0,
          right: 0,
          bottom: 0,
          left: 0
        }}
      >
        <Box
          sx={{ padding: 2, width: maxWidthMatched ? "100%" : "500px" }}
        >
          {resp == null ? (
            <Box sx={{ width: maxWidthMatched ? "100%" : "500px" }} />
          ) : (
            <Recorder streams={resp.body} />
          )}
        </Box>
        <Box
          sx={{ flex: 1, overflow: "hidden" }}
        >
          <BlobInfoQuerier />
        </Box>
      </Box>
    </ThemeProvider>
  );
};
