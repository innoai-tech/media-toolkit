import { format, parseISO } from "date-fns";
import zhCN from "date-fns/locale/zh-CN";
import { BlobInfo, getBlob } from "../client/livestream";
import { Subscribe, useRequest, useStateSubject } from "../utils";
import {
	Box,
	Card,
	CardContent,
	CardHeader,
	Dialog,
	DialogContent,
	Grid,
	IconButton,
	MenuItem,
	MenuList,
	Typography,
	useMediaQuery,
	useTheme,
} from "@mui/material";
import { map } from "@innoai-tech/lodash";
import { Fragment, KeyboardEvent, ReactNode } from "react";
import filesize from "filesize";
import {
	MoreVert,
	OndemandVideoOutlined,
	PhotoOutlined,
} from "@mui/icons-material";
import { TagsInput } from "./TagsInput";
import { PopperWrapper, usePopperController } from "../ui";

const displayLabel = (l: string) => {
	return (
		{
			_media_type: "媒体类型",
			_device_id: "来源设备",
			_size: "文件大小",
		}[l] || l
	);
};

const formatTime = (from: Date | number) => {
	return format(from, "yyyy-MM-dd HH:mm:ss", { locale: zhCN });
};

const BlobCardMedia = ({ mediaType, blobRef, preview }: {
	mediaType: string;
	blobRef: string;
	preview?: boolean;
}) => {
	const getBlob$ = useRequest(getBlob);

	if (mediaType.startsWith("image/")) {
		return (
			<Box>
				<img
					alt={blobRef}
					width={"100%"}
					src={getBlob$.toHref({ ref: blobRef })}
				/>
			</Box>
		);
	}

	if (mediaType.startsWith("video/")) {
		return (
			<Box>
				<video
					src={getBlob$.toHref({ ref: blobRef })}
					width={"100%"}
					preload={preview ? "metadata" : "auto"}
					controls={!preview}
					autoPlay={!preview}
				/>
			</Box>
		);
	}

	return null;
};

const BlobCardMediaPreview = ({ blob }: { blob: BlobInfo }) => {
	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down("md"));
	const mediaType = blob.labels["_media_type"]![0] || "";
	const previewOpen$ = useStateSubject(() => false);

	return (
		<>
			<Box
				sx={{ position: "relative", paddingBottom: `${(100 * 9) / 16}%` }}
				onClick={() => previewOpen$.next(true)}
			>
				<Box
					sx={{
						position: "absolute",
						top: 0,
						right: 0,
						bottom: 0,
						left: 0,
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
					}}
				>
					<BlobCardMedia
						blobRef={blob.ref}
						mediaType={mediaType}
						preview={true}
					/>
				</Box>
				<Box
					sx={{
						position: "absolute",
						top: 0,
						right: 0,
						bottom: 0,
						left: 0,
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						opacity: 0.3,
						fontSize: 36,
						cursor: "pointer",
					}}
				>
					{mediaType.startsWith("image/") ? (
						<PhotoOutlined fontSize={"inherit"} />
					) : (
						<OndemandVideoOutlined fontSize={"inherit"} />
					)}
				</Box>
			</Box>
			<Subscribe value$={previewOpen$}>
				{(opened) => (
					<Dialog
						open={opened}
						fullScreen={fullScreen}
						onClose={() => previewOpen$.next(false)}
					>
						<DialogContent>
							<BlobCardMedia blobRef={blob.ref} mediaType={mediaType} />
						</DialogContent>
					</Dialog>
				)}
			</Subscribe>
		</>
	);
};

export interface BlobInfoCardProps {
	blob: BlobInfo;
	actions: {
		[k: string]: {
			label: ReactNode;
			action: (blob: BlobInfo) => void;
		};
	};
}

export const BlobInfoCardActions = ({ blob, actions }: BlobInfoCardProps) => {
	const [open$, anchorRef] = usePopperController<HTMLButtonElement>();

	return (
		<>
			<IconButton
				ref={anchorRef}
				aria-label="settings"
				onClick={() => open$.next(true)}
			>
				<MoreVert />
			</IconButton>
			<PopperWrapper anchorRef={anchorRef} open$={open$}>
				<MenuList
					id="composition-menu"
					aria-labelledby="composition-button"
					onKeyDown={(event: KeyboardEvent) => {
						if (event.key === "Escape") {
							open$.next(false);
						}
					}}
				>
					{map(
						actions,
						(a, key) => (
							<MenuItem
								key={key}
								onClick={() => {
									a.action(blob);
									open$.next(false);
								}}
							>
								{a.label}
							</MenuItem>
						),
					)}
				</MenuList>
			</PopperWrapper>
		</>
	);
};

export const BlobInfoCard = ({ blob, actions }: BlobInfoCardProps) => {
	return (
		<Card sx={{ width: "100%" }}>
			<CardHeader
				subheader={
					<Typography variant="caption" sx={{ fontSize: 12 }}>
						{formatTime(parseISO(blob.from))}
					</Typography>
				}
				action={<BlobInfoCardActions blob={blob} actions={actions} />}
			/>
			<BlobCardMediaPreview blob={blob} />
			<CardContent>
				{map(blob.labels, (values, label) => {
					if (!label.startsWith("_")) {
						return null;
					}
					return (
						<Box
							key={label}
							sx={{
								display: "flex",
								alignItems: "center",
								justifyContent: "space-between",
							}}
						>
							<Typography
								variant="caption"
								display="block"
								sx={{ fontSize: 10 }}
							>
								{displayLabel(label)}
							</Typography>
							<Typography variant="caption" display="block">
								{values.map((v) => {
									if (label === "_size") {
										return filesize(parseInt(v));
									}
									return v;
								}).join(", ")}
							</Typography>
						</Box>
					);
				})}
				<Box sx={{ p: 1 }} />
				<Grid container={true}>
					{map(
						{
							tag: [],
							...blob.labels,
						},
						(values, label) => {
							if (label.startsWith("_")) {
								return null;
							}
							return (
								<Fragment key={label}>
									<TagsInput blobRef={blob.ref} label={label} values={values} />
								</Fragment>
							);
						},
					)}
				</Grid>
			</CardContent>
		</Card>
	);
};
