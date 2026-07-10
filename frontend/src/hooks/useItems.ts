import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createItem,
  deleteItem,
  fetchCategories,
  fetchItem,
  fetchItems,
  updateItem,
} from '@/services/items';
import type { ItemInput, ItemListParams } from '@/types';

/** Query key factory to keep cache keys consistent across hooks. */
export const itemKeys = {
  all: ['items'] as const,
  lists: () => [...itemKeys.all, 'list'] as const,
  list: (params: ItemListParams) => [...itemKeys.lists(), params] as const,
  details: () => [...itemKeys.all, 'detail'] as const,
  detail: (id: number) => [...itemKeys.details(), id] as const,
  categories: ['items', 'categories'] as const,
};

/**
 * useItems fetches a paginated, category-filtered list of items.
 */
export function useItems(params: ItemListParams) {
  return useQuery({
    queryKey: itemKeys.list(params),
    queryFn: () => fetchItems(params),
    placeholderData: (prev) => prev, // keep previous data while fetching next page
  });
}

/**
 * useItem fetches a single item by ID.
 */
export function useItem(id: number | undefined) {
  return useQuery({
    queryKey: itemKeys.detail(id ?? -1),
    queryFn: () => fetchItem(id!),
    enabled: !!id,
  });
}

/**
 * useCategories fetches distinct item categories for the filter dropdown.
 */
export function useCategories() {
  return useQuery({
    queryKey: itemKeys.categories,
    queryFn: fetchCategories,
    staleTime: 5 * 60 * 1000, // categories change rarely
  });
}

/**
 * useCreateItem mutates the backend and invalidates the item list cache.
 */
export function useCreateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ItemInput) => createItem(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: itemKeys.lists() });
      qc.invalidateQueries({ queryKey: itemKeys.categories });
    },
  });
}

/**
 * useUpdateItem mutates the backend and invalidates the relevant caches.
 */
export function useUpdateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: ItemInput }) => updateItem(id, input),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: itemKeys.lists() });
      qc.invalidateQueries({ queryKey: itemKeys.detail(id) });
      qc.invalidateQueries({ queryKey: itemKeys.categories });
    },
  });
}

/**
 * useDeleteItem soft-deletes an item and invalidates the list cache.
 */
export function useDeleteItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteItem(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: itemKeys.lists() });
      qc.invalidateQueries({ queryKey: itemKeys.categories });
    },
  });
}
