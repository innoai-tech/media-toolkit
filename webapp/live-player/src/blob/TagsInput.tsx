import {
	Box,
	Button,
	Chip,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	TextField,
} from "@mui/material";
import { Add } from "@mui/icons-material";
import {
	StateSubject,
	Subscribe,
	useObservable,
	useObservableEffect,
	useRequest,
	useStateSubject,
} from "../utils";
import { labelBlob, unLabelBlob } from "../client/livestream";
import { useForm } from "react-hook-form";
import { map, trim } from "@innoai-tech/lodash";
import { merge, tap } from "rxjs";

export interface TagInputDialog {
	open$: StateSubject<boolean>;
	onConfirm?: (value: string) => void;
}

export const TagInputDialog = ({ open$, onConfirm }: TagInputDialog) => {
	const { register, handleSubmit } = useForm();
	const opened = useObservable(open$);

	const onSubmit = (data: any) => {
		const v = trim(data.value);
		if (v && onConfirm) {
			onConfirm(v);
		}
	};

	return (
		<Dialog
			open={opened}
			onClose={() => {
				open$.next(false);
			}}
		>
			<form onSubmit={handleSubmit(onSubmit)}>
				<DialogTitle>请输入标签</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus={true}
						fullWidth={true}
						type="text"
						margin="dense"
						variant="standard"
						{...register("value")}
					/>
				</DialogContent>
				<DialogActions>
					<Button type={"submit"}>确定</Button>
				</DialogActions>
			</form>
		</Dialog>
	);
};

export interface TagsInputProps {
	blobRef: string;
	label: string;
	values: string[];
}

export const TagsInput = ({ blobRef, label, values }: TagsInputProps) => {
	const tags$ = useStateSubject(() => values);
	const menuOpen$ = useStateSubject(() => false);

	const labelBlob$ = useRequest(labelBlob);
	const unLabelBlob$ = useRequest(unLabelBlob);

	useObservableEffect(
		() =>
			merge(
				tags$.pipe(
					tap(() => {
						menuOpen$.next(false);
					}),
				),
				labelBlob$.pipe(
					tap((resp) => {
						tags$.next((values) => [...values, resp.config.inputs.value]);
					}),
				),
				unLabelBlob$.pipe(
					tap((resp) => {
						tags$.next(
							(values) => values.filter((v) => v !== resp.config.inputs.value),
						);
					}),
				),
			),
		[],
	);

	return (
		<Box
			sx={{
				display: "flex",
				flexWrap: "wrap",
				listStyle: "none",
				m: -0.2,
				p: 0,
			}}
		>
			<Box key={"__add__"} sx={{ padding: 0.2 }}>
				<Chip
					icon={<Add />}
					label={"标签"}
					size={"small"}
					variant="outlined"
					sx={{ fontSize: 11 }}
					onClick={() => menuOpen$.next(true)}
				/>
			</Box>
			<Subscribe value$={tags$}>
				{(tagValues) => (
					<>
						{map(
							tagValues,
							(value) => (
								<Box key={value || "__add__"} sx={{ padding: 0.2 }}>
									<Chip
										label={value}
										size={"small"}
										variant="outlined"
										sx={{ fontSize: 11 }}
										onDelete={() =>
											unLabelBlob$.next({
												ref: blobRef,
												label,
												value,
											})}
									/>
								</Box>
							),
						)}
					</>
				)}
			</Subscribe>
			<TagInputDialog
				open$={menuOpen$}
				onConfirm={(value) =>
					labelBlob$.next({
						ref: blobRef,
						label,
						value,
					})}
			/>
		</Box>
	);
};
