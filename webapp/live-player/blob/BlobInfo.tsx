import {
  Box,
  Grid, useMediaQuery
} from "@mui/material";
import { formatRFC3339 } from "date-fns";
import { deleteBlob, listBlob } from "webapp/client/livestream";
import { useObservable, useObservableEffect, useRequest } from "webapp/utils";
import { useEffect, useState } from "react";
import { DateTimeRangePicker, DefaultDateRange } from "../../ui";
import { BlobInfoCard } from "./BlobInfoCard";
import { interval, merge, tap } from "rxjs";

const formatTimeRange = (from: Date | number, to: Date | number) => {
  return `${formatRFC3339(from)}..${formatRFC3339(to)}`;
};

export const BlobInfoQuerier = () => {
  const minWidthMatched = useMediaQuery(`(min-width:${256 * 2}px)`);
  const size = 256;

  const listBlob$ = useRequest(listBlob);
  const deleteBlob$ = useRequest(deleteBlob);

  const [timeRange, updateTimeRange] = useState(DefaultDateRange);

  useObservableEffect(() => {
    const fetch = () => listBlob$.next({
      time: formatTimeRange(timeRange[0], timeRange[1])
    });

    fetch();

    return interval(10 * 1000).pipe(tap(fetch));
  }, []);

  const [list, setList] = useState<typeof listBlob.TRespData>([]);

  useObservableEffect(() => {
    return merge(
      listBlob$.pipe(tap((resp) => setList(resp.body))),
      deleteBlob$.pipe(tap((resp) => setList((list) => list.filter((b) => b.ref != resp.config.inputs.ref))))
    );
  }, []);


  return (
    <Box sx={{ display: "flex", flexDirection: "column", width: "100%", height: "100%" }}>
      <Box sx={{ p: 2 }}>
        <DateTimeRangePicker value={timeRange} onValueChange={(dateRange) => {
          console.log(dateRange);
        }} />
      </Box>
      <Box sx={{ flex: 1, overflowY: "auto" }}>
        <Grid container>
          {list.map((b) => (
            <Box
              key={b.ref}
              sx={{
                p: 2,
                gap: 2,
                width: minWidthMatched ? size : "100%"
              }}
            >
              <BlobInfoCard
                blob={b}
                actions={{
                  "delete": {
                    label: "删除",
                    action: (b) => deleteBlob$.next({ ref: b.ref })
                  }
                }}
              />
            </Box>
          ))}
        </Grid>
      </Box>
    </Box>
  );
};
