import { Box, List, ListItem, ListItemButton, ListItemIcon, ListItemText } from "@mui/material";
import { Photo, VideoFile } from "@mui/icons-material";


export interface CameraRecord {
  type: "video" | "picture";
  id: string;
  from: string;
  backup?: string;
  startedAt: string;
  endedAt: string;
}

export const CameraRecordList = ({ list }: { list: CameraRecord[] }) => {
  return (
    <Box>
      <List>
        {list.map((s) => (
          <ListItem
            key={s.id}
          >
            <ListItemIcon>
              {s.type === "video" ? <VideoFile /> : <Photo />}
            </ListItemIcon>
            <ListItemText
              primary={s.backup}
              secondary={s.from}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );
};
