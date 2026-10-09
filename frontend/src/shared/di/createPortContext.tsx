import { createContext, useContext, type ReactNode } from "react";

export function createPortContext<T>(name: string) {
  const Context = createContext<T | null>(null);

  function Provider({ value, children }: { value: T; children: ReactNode }) {
    return <Context.Provider value={value}>{children}</Context.Provider>;
  }

  function usePort(): T {
    const value = useContext(Context);
    if (value === null) {
      throw new Error(`${name} is not provided. Wrap the tree in its provider.`);
    }
    return value;
  }

  return [Provider, usePort] as const;
}
