import { CalendarPicker, CalendarPickerProps } from "@mui/x-date-pickers";
import { cloneElement, ReactElement, useEffect, useRef, useState } from "react";
import { getDate, getHours, getMinutes, getMonth, getSeconds, getYear, set } from "date-fns";
import { Box, Button, DialogActions, Divider, MenuItem, MenuItemProps, MenuList, Stack } from "@mui/material";
import { padStart } from "@innoai-tech/lodash";


const StaticSelect = <T extends any>({
                                       height,
                                       itemHeight,
                                       value,
                                       onChange,
                                       children
                                     }: {
  height: number,
  itemHeight: number,
  value: T;
  onChange: (a: T) => void;
  children: ReactElement<MenuItemProps>[];
}) => {
  const $ulRef = useRef<HTMLUListElement | null>(null);

  useEffect(() => {
    if (!$ulRef.current) {
      return;
    }

    const $ul = $ulRef.current;

    Array.from($ul.children).forEach(($item, index) => {
      if ($item.getAttribute("value") === `${value}`) {
        $ul.scrollTop = itemHeight * index;
      }
    });
  }, [height, value]);

  return (
    <MenuList
      ref={$ulRef}
      dense
      role={"select"}
      sx={{ overflow: "auto", height: height, p: 0 }}
    >
      {children.map((item) =>
        cloneElement(item, {
          selected: value === item.props.value,
          onClick: () => {
            onChange(item.props.value as T);
          }
        })
      )}
      <MenuItem
        key={"__block__"}
        disabled={true}
        sx={{
          height: height - itemHeight
        }}
      />
    </MenuList>
  );
};


export interface DateTimePickerProps extends Omit<CalendarPickerProps<Date | number>, "date" | "onChange"> {
  value: Date | number,
  onValueChange: (value: Date | number) => void;
}

export const DateTimePicker = ({ value, onValueChange, ...otherProps }: DateTimePickerProps) => {
  const [dt, setDt] = useState(() => set(value, { seconds: 0 }));
  return (
    <Stack direction="row" sx={{ alignItems: "stretch" }}>
      <Box>
        <CalendarPicker
          disableHighlightToday={true}
          {...otherProps}
          date={dt}
          onChange={(v) => {
            if (v) {
              const d = Date.parse(v.toString());

              setDt((dt) => {
                return set(dt, {
                  year: getYear(d),
                  month: getMonth(d),
                  date: getDate(d)
                });
              });
            }
          }}
        />
      </Box>
      <Stack>
        <DialogActions>
          <Button
            variant="text"
            onClick={() => {
              onValueChange(dt);
            }}
          >
            确认
          </Button>
        </DialogActions>
        <Stack
          direction="row"
          sx={{
            p: 2,
            pt: 1,
            pl: 0,
            flex: 1,
            maxHeight: 304,
            overflow: "hidden",
            width: 60 * 3,
            "& > [role=select]": { flex: 1 }
          }}
        >
          <StaticSelect
            height={280}
            itemHeight={32}
            value={getHours(dt)}
            onChange={(v) => {
              setDt((dt) => {
                return set(dt, {
                  hours: v
                });
              });
            }}
          >
            {new Array(24).fill(0).map((_, i) => (
              <MenuItem value={i} key={i} sx={{ textAlign: "center" }}>
                <Box sx={{ textAlign: "center", width: "100%" }}>
                  {padStart(`${i}`, 2, "0")}
                </Box>
              </MenuItem>
            ))}
          </StaticSelect>
          <Divider orientation="vertical" flexItem />
          <StaticSelect
            height={280}
            itemHeight={32}
            value={getMinutes(dt)}
            onChange={(v) => {
              setDt((dt) => {
                return set(dt, { minutes: v });
              });
            }}
          >
            {new Array(60).fill(0).map((_, i) => (
              <MenuItem value={i} key={i} sx={{ textAlign: "center" }}>
                <Box sx={{ textAlign: "center", width: "100%" }}>
                  {padStart(`${i}`, 2, "0")}
                </Box>
              </MenuItem>
            ))}
          </StaticSelect>
          <Divider orientation="vertical" flexItem />
          <StaticSelect
            height={280}
            itemHeight={32}
            value={getSeconds(dt)}
            onChange={(v) => {
              setDt((dt) => {
                return set(dt, { seconds: v });
              });
            }}
          >
            {new Array(60).fill(0).map((_, i) => (
              <MenuItem value={i} key={i} sx={{ textAlign: "center" }}>
                <Box sx={{ textAlign: "center", width: "100%" }}>
                  {padStart(`${i}`, 2, "0")}
                </Box>
              </MenuItem>
            ))}
          </StaticSelect>
        </Stack>
      </Stack>
    </Stack>
  );
};
