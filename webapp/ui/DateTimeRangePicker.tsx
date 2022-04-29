import { Box, Grid, TextField } from "@mui/material";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { DateTimePicker, LocalizationProvider } from "@mui/x-date-pickers";
import { addDays } from "date-fns";

export type DateRange = [Date | number, Date | number]

export const DefaultDateRange: DateRange = [addDays(Date.now(), -1), Date.now()];

export const DateTimeRangePicker = ({ value }: { value: DateRange, onValueChange: (value: DateRange) => void }) => {
  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Grid container>
        <DateTimePicker
          renderInput={(props) => <TextField {...props} size={"small"} />}
          value={value[0]}
          maxDate={value[1]}
          onChange={(newValue) => {
            // setValue(newValue);
          }}
        />
        <DateTimePicker
          renderInput={(props) => <TextField {...props} size={"small"} />}
          value={value[1]}
          openTo="day"
          maxDate={Date.now()}
          onChange={() => {
            // setValue(newValue);
          }}
        />
      </Grid>
    </LocalizationProvider>
  );
};