import { useState, KeyboardEvent, SyntheticEvent, useRef, RefObject } from "react";
import { ClickAwayListener, Grow, MenuList, Paper, Popper } from "@mui/material";
import { Builder } from "./DialogBuilder";
import { isFunction } from "@innoai-tech/lodash";

interface MenuListBuilderContext {
  anchorRef: RefObject<HTMLElement>;
  toggle: (open: boolean) => void;
}

export interface MenuListBuilderProps extends Builder<MenuListBuilderContext> {

}

export const MenuListBuilder = ({ content, children }: MenuListBuilderProps) => {
  const [open, toggle] = useState(false);

  const handleListKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab") {
      event.preventDefault();
      toggle(false);
    } else if (event.key === "Escape") {
      toggle(false);
    }
  };

  const anchorRef = useRef<HTMLButtonElement>(null);

  const handleClose = (event: Event | SyntheticEvent) => {
    if (anchorRef.current && anchorRef.current.contains(event.target as HTMLElement)) {
      return;
    }

    toggle(false);
  };

  const ctx = {
    anchorRef,
    toggle
  };

  return <>
    {children(ctx)}
    <Popper
      open={open}
      anchorEl={anchorRef.current}
      placement="bottom-end"
      transition
    >
      {({ TransitionProps, placement }) => (
        <Grow
          {...TransitionProps}
          style={{
            transformOrigin: "right top"
          }}
        >
          <Paper sx={{ zIndex: "tooltip" }}>
            <ClickAwayListener onClickAway={handleClose}>
              <MenuList
                autoFocusItem={open}
                id="composition-menu"
                aria-labelledby="composition-button"
                onKeyDown={handleListKeyDown}
              >
                {isFunction(content) ? content(ctx) : content}
              </MenuList>
            </ClickAwayListener>
          </Paper>
        </Grow>
      )}
    </Popper>
  </>;
};