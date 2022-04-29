import { Box, Container, createTheme, CssBaseline, Paper, ThemeProvider } from "@mui/material";
import { formatISO, addMinutes } from "date-fns";
import { Recorder, Source } from "./Recorder";
import { CameraRecord, CameraRecordList } from "./CameraRecordList";
import { useSWRWith } from "../client/client";
import { listStreams } from "../client/livestream";

const theme = createTheme();

const records: CameraRecord[] = new Array(5).fill(0).map((v, i) => ({
  id: `${i}`,
  from: `test_${(100 * Math.random()).toFixed(0)}`,
  name: `测试源_${i}`,
  type: "video",
  startedAt: formatISO(Date.now()),
  endedAt: formatISO(addMinutes(Date.now(), 2))
}));

export const App = () => {
  const { data } = useSWRWith(listStreams, undefined);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container
        component="main"
        maxWidth="sm"
        sx={{
          display: "flex",
          flexDirection: "column",
          height: "100vh",
          overflow: "hidden"
        }}
      >
        <Box
          sx={{ padding: 2 }}
        >
          {data == null ? (
            <Box />
          ) : (
            <Recorder streams={data.data} />
          )}
        </Box>
        <Box
          sx={{ flex: 1, overflow: "auto" }}
        >
          <CameraRecordList list={records} />
        </Box>
      </Container>
    </ThemeProvider>
  );
};
