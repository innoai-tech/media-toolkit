import { Box, Container, createTheme, CssBaseline, Paper, ThemeProvider } from "@mui/material";
import { formatISO, addMinutes } from "date-fns";
import { Recorder, Source } from "./Recorder";
import { CameraRecord, CameraRecordList } from "./CameraRecordList";
// @ts-ignore
import flv from "samples/bun33s.flv?url";

const theme = createTheme();

const sources: Source[] = new Array(100).fill(0).map((v, i) => ({
  id: `test_${i}`,
  name: `测试源_${i}`,
  src: flv
}));

const records: CameraRecord[] = new Array(100).fill(0).map((v, i) => ({
  id: `${i}`,
  from: `test_${(100 * Math.random()).toFixed(0)}`,
  name: `测试源_${i}`,
  type: "video",
  startedAt: formatISO(Date.now()),
  endedAt: formatISO(addMinutes(Date.now(), 2))
}));

export const App = () => {
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
          <Recorder
            sources={sources}
          />
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
