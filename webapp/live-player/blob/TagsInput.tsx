import {
  Box, Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField
} from "@mui/material";
import { Add } from "@mui/icons-material";
import { useObservableEffect, useRequest } from "webapp/utils";
import { labelBlob, unLabelBlob } from "webapp/client/livestream";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { trim } from "@innoai-tech/lodash";
import { merge, tap } from "rxjs";

export interface TagInputDialog {
  open: boolean,
  onConfirm?: (value: string) => void
  onCancel?: () => void
}

export const TagInputDialog = ({ open, onConfirm, onCancel }: TagInputDialog) => {
  const { register, handleSubmit } = useForm();

  const onSubmit = (data: any) => {
    const v = trim(data.value);
    if (v && onConfirm) {
      onConfirm(v);
    }
  };

  return (
    <Dialog open={open} onClose={onCancel}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <DialogTitle>请输入标签</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            type="text"
            margin="dense"
            variant="standard"
            {...register("value")}
          />
        </DialogContent>
        <DialogActions>
          <Button type={"submit"}>
            确定
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};


export interface TagsInputProps {
  blobRef: string
  label: string,
  values: string[],
}

export const TagsInput = ({
                            blobRef,
                            label,
                            values
                          }: TagsInputProps) => {

  const [vals, updateValues] = useState(() => values);

  const labelBlob$ = useRequest(labelBlob);
  const unLabelBlob$ = useRequest(unLabelBlob);

  const [open, toggleInput] = useState(false);

  useObservableEffect(() => merge(
    labelBlob$.pipe(tap((resp) => {
      updateValues((values) => [...values, resp.config.inputs.value]);
      toggleInput(false);
    })),
    unLabelBlob$.pipe(tap((resp) => {
      updateValues((values) => values.filter((v) => v != resp.config.inputs.value));
      toggleInput(false);
    }))
  ), []);

  return (
    <Box sx={{
      display: "flex",
      flexWrap: "wrap",
      listStyle: "none",
      m: -0.2,
      p: 0
    }}>
      <Box key={"__add__"} sx={{ padding: 0.2 }}>
        <Chip
          icon={<Add />}
          label={"标签"}
          size={"small"}
          variant="outlined"
          sx={{ fontSize: 11 }}
          onClick={() => toggleInput(true)}
        />
      </Box>
      {[...vals].map((value) => {
        return (
          <Box key={value || "__add__"} sx={{ padding: 0.2 }}>
            <Chip
              label={value}
              size={"small"}
              variant="outlined"
              sx={{ fontSize: 11 }}
              onDelete={() => unLabelBlob$.next({
                ref: blobRef,
                label,
                value
              })}
            />
          </Box>
        );
      })}
      <TagInputDialog
        open={open}
        onCancel={() => toggleInput(false)}
        onConfirm={(value) => labelBlob$.next({
          ref: blobRef,
          label,
          value
        })}
      />
    </Box>
  );
};