import { useEffect, useState } from 'react';

/**
 * useDebounce returns a debounced version of the given value after the
 * specified delay. Used for search inputs to avoid firing a request on
 * every keystroke.
 *
 * @param value The rapidly-changing value to debounce.
 * @param delay Minimum delay in ms (default 300ms per the spec).
 */
export function useDebounce<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(handler);
  }, [value, delay]);

  return debounced;
}
