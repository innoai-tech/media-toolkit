import { createContext, ReactNode, useContext } from "react";
import { liveStreamStatus } from "webapp/client/livestream";
import { useObservable, useObservableEffect, useRequest } from "webapp/utils";
import { interval, tap } from "rxjs";

const StreamStatusContext = createContext({} as (typeof liveStreamStatus.TRespData));

export const useStreamStatus = () => useContext(StreamStatusContext);

const defaultStatus = { active: false, observers: {} };

export const StreamStatusProvider = ({ id, children }: { id: string, children?: ReactNode }) => {
  const request$ = useRequest(liveStreamStatus);
  const resp = useObservable(request$);

  // useObservableEffect(() => {
  //   const fetch = () => request$.next({ id });
  //
  //   return interval(1000).pipe(tap(fetch));
  // }, []);

  return (
    <StreamStatusContext.Provider value={
      resp?.body || defaultStatus
    }>
      {children}
    </StreamStatusContext.Provider>
  );
};
