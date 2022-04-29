import { ReactNode, useState } from "react";
import { Dialog, DialogProps } from "@mui/material";
import { isFunction } from "@innoai-tech/lodash";

export interface Builder<BuilderContext> {
  content: ReactNode | ((ctx: BuilderContext) => ReactNode);
  children: (ctx: BuilderContext) => ReactNode;
}

interface OpenContext {
  open: boolean;
  toggle: (open: boolean) => void;
}

export interface DialogBuilderProps extends Omit<DialogProps, "children" | "open">, Builder<OpenContext> {
}

export const DialogBuilder = ({ content, children, ...props }: DialogBuilderProps) => {
  const [open, toggle] = useState(false);

  return (
    <>
      {children({ open, toggle })}
      <Dialog
        {...props}
        open={open}
        onClose={(...args) => {
          toggle(false);
          props.onClose && props.onClose(...args);
        }}
      >
        {isFunction(content) ? content({ open, toggle }) : content}
      </Dialog>
    </>
  );
};