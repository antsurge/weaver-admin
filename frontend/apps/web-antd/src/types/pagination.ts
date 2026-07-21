export const DEFAULT_PAGE_SIZE = 20;

export const PAGE_SIZES = [10, 20, 50, 100,200];

export interface PaginationParams {
  currentPage?:number;
  pageSize?:number;
}

export interface PaginationResult<T> {
  items: T[];
  total: number;
}

export interface AllResult<T> {
  items: T[];
}
