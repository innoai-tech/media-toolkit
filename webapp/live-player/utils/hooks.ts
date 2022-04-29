import { useEffect } from "react";
import { merge, Observable } from "rxjs";

export type ArrayOrNone<T> = T | T[] | undefined;

export const useObservableEffect = (effect: () => ArrayOrNone<Observable<any>>, deps: any[] = []) => {
  useEffect(() => {
    const ob = effect();
    if (!ob) {
      return;
    }
    const sub = merge(...([] as Array<Observable<any>>).concat(ob)).subscribe();
    return sub.unsubscribe;
  }, deps);
};