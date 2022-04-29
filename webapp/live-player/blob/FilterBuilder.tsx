import { Chip, Divider, InputBase, Menu, MenuItem, MenuList, Select, Stack, Typography } from "@mui/material";
import { DoubleArrow } from "@mui/icons-material";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { addDays, format } from "date-fns";
import zhCN from "date-fns/locale/zh-CN";
import { map, size, isEmpty, uniq, omit, trim } from "@innoai-tech/lodash";
import { StateSubject, Subscribe, useObservable, useObservableEffect, useStateSubject } from "webapp/utils";
import { PopperWrapper, usePopperController } from "webapp/ui";
import { DateTimePicker } from "./DateTimePicker";
import { tap } from "rxjs";
import { Fragment, useMemo } from "react";

export type DateRange = [Date | number, Date | number];

export const DefaultDateRange: DateRange = [addDays(Date.now(), -1), Date.now()];

const formatDateTime = (date: Date | number) => format(date, "yyyy-MM-dd HH:mm:ss", { locale: zhCN });

export interface LabelMeta {
  display: string;
  values?: { [key: string]: string };
  multiple?: boolean;
}

export interface FilterValue {
  time: DateRange;
  filter: { [key: string]: string[] };
}

export interface FilterBuilderProps {
  labels: { [k: string]: LabelMeta };
  filterValue$: StateSubject<FilterValue>;
}

export const FilterBuilder = ({ labels, filterValue$ }: FilterBuilderProps) => {
  const [openFrom$, anchorRefFrom] = usePopperController<HTMLInputElement>();
  const [openTo$, anchorRefTo] = usePopperController<HTMLInputElement>();
  const [openFilterKey$, filterPopperRef] = usePopperController<HTMLInputElement>();

  const filterKey$ = useStateSubject(() => "");

  const displayKey = (key: string) => {
    return labels[key]?.display || key;
  };

  useObservableEffect(() => {
    return filterKey$.pipe(
      tap((k) => {
        if (!!k) {
          openFilterKey$.next(false);
        }
      })
    );
  }, []);

  const { delFilter, addFilter } = useMemo(() => {
    return {
      delFilter: (k: string, v: string) => {
        filterValue$.next((filterValue) => {
          const nextValues = (filterValue.filter[k] ?? []).filter((cv) => cv !== v);
          return {
            ...filterValue,
            filter: isEmpty(nextValues)
              ? omit(filterValue.filter, k)
              : {
                ...filterValue.filter,
                [k]: nextValues
              }
          };
        });
      },
      addFilter: (k: string, v: string) => {
        // clear key
        filterKey$.next("");

        filterValue$.next((filterValue) => ({
          ...filterValue,
          filter: {
            ...filterValue.filter,
            [k]: uniq([...(filterValue.filter[k] ?? []), v])
          }
        }));
      }
    };
  }, []);

  const filterValue = useObservable(filterValue$);

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={zhCN}>
      {size(filterValue.filter) > 0 && (
        <Stack direction={"row"} gap={1} sx={{ pl: 1 }}>
          {map(filterValue.filter, (values, key) => (
            <Fragment key={key}>
              {map(values, (v) => (
                <Chip
                  key={v}
                  label={
                    <Stack gap={0.5} direction={"row"}>
                      <Typography variant={"caption"}>{displayKey(key)}</Typography>
                      <Divider orientation="vertical" />
                      <Typography variant={"caption"}>{v}</Typography>
                    </Stack>
                  }
                  onDelete={() => {
                    delFilter(key, v);
                  }}
                />
              ))}
            </Fragment>
          ))}
        </Stack>
      )}
      <Subscribe value$={filterKey$}>
        {(key) =>
          key ? (
            <>
              <Chip
                sx={{
                  borderTopRightRadius: 0,
                  borderBottomRightRadius: 0
                }}
                label={<Typography variant={"caption"}>{displayKey(key)}</Typography>}
              />
            </>
          ) : null
        }
      </Subscribe>
      <InputBase
        sx={{ flex: 1, pl: 1 }}
        placeholder="输入筛选条件"
        fullWidth
        ref={filterPopperRef}
        onFocus={(e) => {
          // only direct focus can trigger
          if (!e.relatedTarget && isEmpty(filterKey$.value)) {
            openFilterKey$.next(true);
          }
        }}
        onKeyDown={(e) => {
          const target$ = e.target as HTMLInputElement;

          const inputValue = trim(target$.value);
          if (e.code === "Enter") {
            if (inputValue && filterKey$.value) {
              addFilter(filterKey$.value, inputValue);
              target$.value = "";
            }
          }
          if (inputValue === "" && e.code === "Backspace") {
            filterKey$.next("");
            target$.blur();
            target$.focus();
          }
        }}
      />
      <Subscribe value$={filterKey$}>
        {(key) => {
          return key && labels[key].values ? (
            <Menu
              open={true}
              anchorEl={filterPopperRef.current}
              onClose={() => {
                openFilterKey$.next(false);
              }}
              anchorOrigin={{
                vertical: "top",
                horizontal: "left"
              }}
            >
              {map(labels[filterKey$.value].values, (label, value) => {
                return (
                  <MenuItem
                    key={value}
                    value={value}
                    dense
                    onClick={() => {
                      addFilter(key, value);
                    }}
                    onKeyDown={(e) => {
                      if (e.code === "Enter") {
                        addFilter(key, value);
                      }
                    }}
                  >
                    {label}
                  </MenuItem>
                );
              })}
            </Menu>
          ) : null;
        }}
      </Subscribe>
      <Subscribe value$={openFilterKey$}>
        {(opened) => {
          return (
            <Menu
              open={opened}
              anchorEl={filterPopperRef.current}
              onClose={() => {
                openFilterKey$.next(false);
              }}
              anchorOrigin={{
                vertical: "top",
                horizontal: "left"
              }}
            >
              {map(labels, (labelMeta, key) => {
                if (!labelMeta.multiple && !isEmpty(filterValue.filter[key])) {
                  return null;
                }

                return (
                  <MenuItem
                    key={key}
                    value={key}
                    dense
                    onClick={() => {
                      filterKey$.next(key);
                    }}
                    onKeyDown={(e) => {
                      if (e.code === "Enter") {
                        filterKey$.next(key);
                      }
                    }}
                  >
                    {labelMeta.display}
                  </MenuItem>
                );
              })}
            </Menu>
          );
        }}
      </Subscribe>
      <Divider orientation="vertical" />
      <InputBase
        ref={anchorRefFrom}
        onFocus={() => openFrom$.next(true)}
        sx={{ "& input": { textAlign: "center" } }}
        value={formatDateTime(filterValue.time[0])}
      />
      <DoubleArrow fontSize={"inherit"} color={"inherit"} />
      <InputBase
        ref={anchorRefTo}
        onFocus={() => openTo$.next(true)}
        sx={{ "& input": { textAlign: "center" } }}
        value={formatDateTime(filterValue.time[1])}
      />
      <PopperWrapper open$={openFrom$} anchorRef={anchorRefFrom}>
        <DateTimePicker
          value={filterValue.time[0]}
          onValueChange={(value) => {
            filterValue$.next((filterValue) => ({
              ...filterValue,
              time: [value, filterValue.time[1]]
            }));
            openFrom$.next(false);
          }}
        />
      </PopperWrapper>
      <PopperWrapper open$={openTo$} anchorRef={anchorRefTo}>
        <DateTimePicker
          value={filterValue.time[1]}
          maxDate={Date.now()}
          onValueChange={(value) => {
            filterValue$.next((filterValue) => ({
              ...filterValue,
              time: [filterValue.time[0], value]
            }));
            openTo$.next(false);
          }}
        />
      </PopperWrapper>
    </LocalizationProvider>
  );
};
