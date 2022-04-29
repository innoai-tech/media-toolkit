import { useEffect, useState } from "react";
import { merge, Observable, tap } from "rxjs";

export type ArrayOrNone<T> = T | T[] | undefined;


export const useObservableEffect = (effect: () => ArrayOrNone<Observable<any>>, deps: any[] = []) => {
  useEffect(() => {
    const ob = effect();
    if (!ob) {
      return;
    }
    const sub = merge(...([] as Array<Observable<any>>).concat(ob)).subscribe();
    return () => sub.unsubscribe();
  }, deps);
};

export const useObservable = <T extends any>(ob$: Observable<T>): T | null => {
  const [s, up] = useState(() => (ob$ as any).value);
  useObservableEffect(() => ob$.pipe(tap((resp) => up(resp))), []);
  return s;
};
