import { ReactNode, RefObject, SyntheticEvent, useRef } from "react";
import {
	Box,
	ClickAwayListener,
	Grow,
	Paper,
	Popper,
	PopperProps,
} from "@mui/material";
import { StateSubject, Subscribe, useStateSubject } from "../utils";

export const usePopperController = <T extends HTMLElement>() => {
	const anchorRef = useRef<T>(null);
	const open$ = useStateSubject(false);
	return [open$, anchorRef] as const;
};

export interface PopperWrapperProps<T extends HTMLElement>
	extends Omit<PopperProps, "anchorEl" | "open" | "children"> {
	open$: StateSubject<boolean>;
	children: ReactNode;
	anchorRef: RefObject<T>;
}

export const PopperWrapper = <T extends HTMLElement>({
	open$,
	anchorRef,
	children,
	transition = true,
	...otherProps
}: PopperWrapperProps<T>) => {
	const c = (
		<Paper sx={{ zIndex: "tooltip" }}>
			<ClickAwayListener
				onClickAway={(event: Event | SyntheticEvent) => {
					if (
						anchorRef.current &&
						anchorRef.current.contains(event.target as HTMLElement)
					) {
						return;
					}
					open$.next(false);
				}}
			>
				<Box>{children}</Box>
			</ClickAwayListener>
		</Paper>
	);

	return (
		<Subscribe value$={open$}>
			{(opened) => (
				<Popper
					open={opened}
					anchorEl={anchorRef.current}
					transition={transition}
					{...otherProps}
				>
					{transition
						? ({ TransitionProps, placement }) => (
								<Grow
									{...TransitionProps}
									style={{
										transformOrigin: placement.includes("bottom")
											? "top"
											: "bottom",
									}}
								>
									{c}
								</Grow>
						  )
						: c}
				</Popper>
			)}
		</Subscribe>
	);
};
