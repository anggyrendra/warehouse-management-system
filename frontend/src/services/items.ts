import { request, requestFull } from './api';
import type {
  Item,
  ItemInput,
  ItemListParams,
  Location,
  StockReceiveInput,
  StockWithDetail,
} from '@/types';

/**
 * Build a query string from ItemListParams, omitting empty/undefined values.
 */
function buildQuery(params: ItemListParams): string {
  const sp = new URLSearchParams();
  if (params.category) sp.set('category', params.category);
  if (params.page) sp.set('page', String(params.page));
  if (params.limit) sp.set('limit', String(params.limit));
  // NOTE: search is not a backend parameter — it is handled client-side
  // because the backend List endpoint filters by category only. We keep
  // the param in the type for future server-side search support.
  const qs = sp.toString();
  return qs ? `?${qs}` : '';
}

/** List items with optional category filter and pagination. */
export async function fetchItems(params: ItemListParams = {}) {
  return requestFull<Item[]>(`/items${buildQuery(params)}`);
}

/** Fetch a single item by ID. */
export async function fetchItem(id: number): Promise<Item> {
  return request<Item>(`/items/${id}`);
}

/** Create a new item. */
export async function createItem(input: ItemInput): Promise<Item> {
  return request<Item>('/items', { method: 'POST', body: input });
}

/** Update an existing item. */
export async function updateItem(id: number, input: ItemInput): Promise<Item> {
  return request<Item>(`/items/${id}`, { method: 'PUT', body: input });
}

/** Soft-delete an item. */
export async function deleteItem(id: number): Promise<void> {
  await request<null>(`/items/${id}`, { method: 'DELETE' });
}

/** Fetch all locations (for the stock-receive form / future use). */
export async function fetchLocations(): Promise<Location[]> {
  return request<Location[]>('/locations');
}

/** Fetch distinct item categories (for the filter dropdown). */
export async function fetchCategories(): Promise<string[]> {
  return request<string[]>('/items-categories');
}

/** Receive stock into a location. */
export async function receiveStock(input: StockReceiveInput): Promise<StockWithDetail> {
  return request<StockWithDetail>('/stock/receive', { method: 'POST', body: input });
}

/** Check stock for an item across all locations. */
export async function fetchStockByItem(itemId: number): Promise<StockWithDetail[]> {
  return request<StockWithDetail[]>(`/stock/${itemId}`);
}
