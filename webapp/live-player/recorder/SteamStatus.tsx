import { createContext, ReactNode, useContext, useEffect } from "react";
import { liveStreamStatus } from "webapp/client/livestream";
import { StateSubject, useObservableEffect, useRequest, useStateSubject } from "webapp/utils";
import { tap } from "rxjs";

type StreamStatus = typeof liveStreamStatus.TRespData & {
  time?: Date;
};

const StreamStatusContext = createContext(
  {} as {
    status$: StateSubject<StreamStatus>;
  }
);

export const useStreamStatus = () => useContext(StreamStatusContext).status$;

export const StreamStatusProvider = ({
  id,
  children,
}: {
  id: string;
  children?: ReactNode;
}) => {
  const request$ = useRequest(liveStreamStatus);
  const status$ = useStateSubject(() => ({ active: false, observers: {} }));

  useObservableEffect(
    () => request$.pipe(tap((resp) => status$.next(resp.body))),
    []
  );

  useEffect(() => {
    request$.next({ id });
  }, [id]);

  return (
    <StreamStatusContext.Provider
      value={{
        status$,
      }}
    >
      {children}
    </StreamStatusContext.Provider>
  );
};
