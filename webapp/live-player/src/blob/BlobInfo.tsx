import {
	Box,
	Divider,
	Grid,
	IconButton,
	Paper,
	Stack,
	useMediaQuery,
} from "@mui/material";
import { Sync, Download } from "@mui/icons-material";
import { formatRFC3339, parseISO } from "date-fns";
import {
	BlobInfo,
	deleteBlob,
	exportDataset,
	listBlob,
} from "../client/livestream";
import {
	formatLabelQL,
	parseLabelQL,
	stringifySearch,
	Subscribe,
	useObservableEffect,
	useRequest,
	useStateSubject,
} from "../utils";
import { useMemo } from "react";
import { BlobInfoCard } from "./BlobInfoCard";
import { distinctUntilChanged, interval, merge, tap } from "rxjs";
import {
	DefaultDateRange,
	FilterBuilder,
	FilterValue,
	LabelMeta,
} from "./FilterBuilder";
import { isEmpty, isEqual, map } from "@innoai-tech/lodash";
import { useLocation, useNavigate } from "react-router";

const formatTimeRange = (from: Date | number, to: Date | number) => {
	return `${formatRFC3339(from)}..${formatRFC3339(to)}`;
};

const parseTimeRange = (
	timeRange: string,
): [Date | number, Date | number] | null => {
	try {
		const values = timeRange.split("..").map((t) => parseISO(t));
		return values.length === 2
			? (values as [Date | number, Date | number])
			: null;
	} catch (_) {
		return null;
	}
};

const labels: { [k: string]: LabelMeta } = {
	_media_type: {
		display: "媒体类型",
		values: {
			"image/jpeg": "图片",
			"video/mp4": "视频",
		},
	},
	tag: {
		display: "标签",
		multiple: true,
	},
};

export const BlobInfoQuerier = () => {
	const exportDataset$ = useRequest(exportDataset);
	const listBlob$ = useRequest(listBlob);
	const deleteBlob$ = useRequest(deleteBlob);
	const list$ = useStateSubject<BlobInfo[]>([]);
	const location = useLocation();
	const navigate = useNavigate();

	const valuesFromSearch = useMemo(() => {
		const sp = new URLSearchParams(location.search);

		return {
			time: parseTimeRange(sp.get("time") || "") || DefaultDateRange,
			filter: parseLabelQL(sp.get("filter") || "") as any,
		};
	}, [location.search]);

	console.log(valuesFromSearch);

	const filterValue$ = useStateSubject<FilterValue>(
		() =>
			isEmpty(valuesFromSearch)
				? {
						time: DefaultDateRange,
						filter: {},
				  }
				: valuesFromSearch,
	);

	const fetch = useMemo(() => {
		return (filterValue: FilterValue) => {
			listBlob$.next({
				time: formatTimeRange(filterValue.time[0], filterValue.time[1]),
				filter: formatLabelQL(filterValue.filter),
			});
		};
	}, []);

	useObservableEffect(() => {
		fetch(filterValue$.value);
		return interval(10 * 1000).pipe(tap(() => fetch(filterValue$.value)));
	}, []);

	useObservableEffect(() => {
		return merge(
			filterValue$.pipe(
				distinctUntilChanged(isEqual),
				tap((nextFilterValue: FilterValue) => {
					fetch(nextFilterValue);

					navigate(
						{
							search: stringifySearch({
								time: [
									formatTimeRange(
										nextFilterValue.time[0],
										nextFilterValue.time[1],
									),
								],
								filter: [formatLabelQL(nextFilterValue.filter)],
							}),
						},
						{ replace: true },
					);
				}),
			),
			listBlob$.pipe(tap((resp) => list$.next(resp.body))),
			deleteBlob$.pipe(
				tap(
					(resp) =>
						list$.next(
							(list) => list.filter((b) => b.ref !== resp.config.inputs.ref),
						),
				),
			),
		);
	}, []);

	const size = 256;
	const minWidthMatched = useMediaQuery(`(min-width:${size * 2}px)`);

	return (
		<Stack gap={3} sx={{ width: "100%", height: "100%" }}>
			<Paper sx={{ maxWidth: 1200 }}>
				<Stack
					direction={"row"}
					gap={0.5}
					sx={{
						alignItems: "center",
						padding: 0.5,
						borderRadius: 2,
					}}
				>
					<FilterBuilder labels={labels} filterValue$={filterValue$} />
					<Divider orientation="vertical" sx={{ w: 1 }} />
					<Subscribe value$={listBlob$.requesting$}>
						{(requesting) => (
							<IconButton
								disabled={requesting}
								onClick={() => fetch(filterValue$.value)}
							>
								<Sync />
							</IconButton>
						)}
					</Subscribe>
					<Subscribe value$={filterValue$}>
						{(filterValue) => (
							<IconButton
								component={"a"}
								target={"_blank"}
								href={exportDataset$.toHref({
									time: formatTimeRange(
										filterValue.time[0],
										filterValue.time[1],
									),
									filter: formatLabelQL(filterValue.filter),
								})}
							>
								<Download />
							</IconButton>
						)}
					</Subscribe>
				</Stack>
			</Paper>
			<Box sx={{ flex: 1, overflowY: "auto" }}>
				<Grid container={true} gap={2}>
					<Subscribe value$={list$}>
						{(list) => (
							<>
								{map(
									list,
									(b) => (
										<Box
											key={b.ref}
											sx={{
												width: minWidthMatched ? size : "100%",
											}}
										>
											<BlobInfoCard
												blob={b}
												actions={{
													delete: {
														label: "删除",
														action: (b) => deleteBlob$.next({ ref: b.ref }),
													},
												}}
											/>
										</Box>
									),
								)}
							</>
						)}
					</Subscribe>
				</Grid>
			</Box>
		</Stack>
	);
};
