import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchLocations, fetchStockByItem, receiveStock } from '@/services/items';
import type { StockReceiveInput } from '@/types';

export const stockKeys = {
  all: ['stock'] as const,
  byItem: (itemId: number) => [...stockKeys.all, 'item', itemId] as const,
  locations: ['locations'] as const,
};

/** useStockByItem fetches stock rows for a given item across all locations. */
export function useStockByItem(itemId: number | undefined) {
  return useQuery({
    queryKey: stockKeys.byItem(itemId ?? -1),
    queryFn: () => fetchStockByItem(itemId!),
    enabled: !!itemId,
  });
}

/** useLocations fetches all warehouse locations. */
export function useLocations() {
  return useQuery({
    queryKey: stockKeys.locations,
    queryFn: fetchLocations,
    staleTime: 5 * 60 * 1000,
  });
}

/** useReceiveStock mutates stock and invalidates the stock cache for the item. */
export function useReceiveStock() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: StockReceiveInput) => receiveStock(input),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: stockKeys.byItem(input.item_id) });
    },
  });
}
