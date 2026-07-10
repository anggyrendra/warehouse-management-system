// Types matching the backend API response envelope and domain entities.

/** Standard API success response envelope. */
export interface SuccessResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: PaginationMeta;
}

/** Standard API error response envelope. */
export interface ErrorResponse {
  success: false;
  message: string;
  errors?: FieldError[];
}

/** A single field-level validation error. */
export interface FieldError {
  field: string;
  reason: string;
}

/** Pagination metadata returned by list endpoints. */
export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
}

/** Warehouse item (master data). */
export interface Item {
  id: number;
  sku: string;
  name: string;
  category: string;
  unit: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

/** Request payload for creating/updating an item. */
export interface ItemInput {
  sku: string;
  name: string;
  category: string;
  unit: string;
}

/** Warehouse location (rack/zone). */
export interface Location {
  id: number;
  code: string;
  zone: string;
  type: string;
  created_at: string;
}

/** Stock row with joined item + location details. */
export interface StockWithDetail {
  id: number;
  item_id: number;
  item_sku: string;
  item_name: string;
  location_id: number;
  location_code: string;
  zone: string;
  qty: number;
  updated_at: string;
}

/** Stock receive request payload. */
export interface StockReceiveInput {
  item_id: number;
  location_id: number;
  qty: number;
}

/** Parameters for the item list query. */
export interface ItemListParams {
  category?: string;
  search?: string;
  page?: number;
  limit?: number;
}
