const BASE = '/api/v1';

let accessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

export function setToken(token: string | null) { accessToken = token; }
export function getToken(): string | null { return accessToken; }

async function refreshAccessToken(): Promise<string | null> {
  try {
    const res = await fetch(`${BASE}/auth/refresh`, { method: 'POST', credentials: 'include' });
    if (!res.ok) return null;
    const data = await res.json();
    accessToken = data.access_token;
    return accessToken;
  } catch { return null; }
}

async function request<T = any>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${BASE}${path}`;
  const headers: Record<string, string> = {};
  if (!(options.body instanceof FormData)) headers['Content-Type'] = 'application/json';
  Object.assign(headers, (options.headers as Record<string, string>) || {});
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

  let res = await fetch(url, { ...options, headers, credentials: 'include' });
  if (res.status === 401 && !path.includes('/auth/')) {
    if (!refreshPromise) refreshPromise = refreshAccessToken();
    const newToken = await refreshPromise;
    refreshPromise = null;
    if (newToken) {
      headers['Authorization'] = `Bearer ${newToken}`;
      res = await fetch(url, { ...options, headers, credentials: 'include' });
    }
  }
  const data = await res.json();
  if (!res.ok) {
    const err = data.error || { code: 'UNKNOWN', message: 'Unknown error' };
    throw new ApiError(res.status, err.code, err.message);
  }
  return data as T;
}

export class ApiError extends Error {
  status: number; code: string;
  constructor(status: number, code: string, message: string) {
    super(message); this.status = status; this.code = code; this.name = 'ApiError';
  }
}

// Auth
export const auth = {
  login: (login: string, password: string) =>
    request<{ access_token: string; user: { id: string; username: string; email: string } }>('/auth/login', { method: 'POST', body: JSON.stringify({ login, password }) }),
  refresh: () => request<{ access_token: string }>('/auth/refresh', { method: 'POST' }),
  logout: () => request('/auth/logout', { method: 'POST' }),
  me: () => request<{ id: string; username: string }>('/admin/me'),
};

// Setup
export const setup = {
  status: () => request<{ setup_required: boolean }>('/setup/status'),
  initialize: (data: { username: string; email: string; password: string; site_title: string; site_subtitle?: string; site_desc?: string }) =>
    request<{ access_token: string; user: { id: string; username: string; email: string } }>('/setup/initialize', { method: 'POST', body: JSON.stringify(data) }),
};

// Posts
export interface PostData {
  id: string; title: string; slug: string; content: string; content_html: string;
  summary: string; category: string; tags: string; is_pinned: boolean; is_draft: boolean;
  is_private: boolean; published_at: string | null; created_at: string; updated_at: string; deleted_at: string | null;
}

export interface PublicPostItem {
  id: string; title: string; slug: string; summary: string; category: string;
  tags: string; is_pinned: boolean; published_at: string | null; created_at: string;
}

export const posts = {
  list: (params?: Record<string, string | number>) =>
    request('/admin/posts' + qs(params)),
  get: (id: string) => request<PostData>(`/admin/posts/${id}`),
  create: (data: Record<string, unknown>) => request<PostData>('/admin/posts', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Record<string, unknown>) => request<PostData>(`/admin/posts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request(`/admin/posts/${id}`, { method: 'DELETE' }),
  publish: (id: string, publish: boolean) => request(`/admin/posts/${id}/publish`, { method: 'PUT', body: JSON.stringify({ publish }) }),
  pin: (id: string, pin: boolean) => request(`/admin/posts/${id}/pin`, { method: 'PUT', body: JSON.stringify({ pin }) }),
  dashboard: () => request<{ total_posts: number; published_posts: number; draft_posts: number; treehole_posts: number }>('/admin/dashboard'),
};

// Public frontend
export const front = {
  posts: (params?: Record<string, string | number>) =>
    request<{ posts: PublicPostItem[]; total: number; page: number; per_page: number }>('/posts' + qs(params)),
  postBySlug: (slug: string) => request<PostData>(`/posts/${slug}`),
  search: (q: string, page = 1) =>
    request<{ posts: { id: string; title: string; slug: string; summary: string; category: string; published_at: string | null }[]; total: number }>(`/posts/search?q=${encodeURIComponent(q)}&page=${page}`),
  archive: () => request<{ archive: { year: string; posts: { id: string; title: string; slug: string; category: string; published_at: string | null }[] }[] }>('/posts/archive'),
  page: (slug: string) => request<{ slug: string; content: string }>(`/pages/${slug}`),
};

// Comments
export interface CommentData {
  id: string; post_id: string; parent_id: string; author_name: string;
  author_email?: string; author_url?: string; content: string; status: string; created_at: string;
}

export const comments = {
  getByPost: (slug: string) => request<{ comments: CommentData[] }>(`/posts/${slug}/comments`),
  create: (slug: string, data: { author_name: string; author_email: string; author_url?: string; content: string; parent_id?: string }) =>
    request<CommentData>(`/posts/${slug}/comments`, { method: 'POST', body: JSON.stringify(data) }),
  list: (params?: Record<string, string | number>) => request<{ comments: CommentData[]; total: number; page: number; per_page: number }>('/admin/comments' + qs(params)),
  updateStatus: (id: string, status: string) => request(`/admin/comments/${id}`, { method: 'PUT', body: JSON.stringify({ status }) }),
  delete: (id: string) => request(`/admin/comments/${id}`, { method: 'DELETE' }),
};

// Admin pages
export const pages = {
  get: (slug: string) => request<{ slug: string; content: string }>(`/pages/${slug}`),
  update: (slug: string, content: string) => request<{ slug: string; content: string }>(`/admin/pages/${slug}`, { method: 'PUT', body: JSON.stringify({ content }) }),
};

// Media
export interface MediaData {
  id: string; file_name: string; original_name: string; file_path: string;
  file_size: number; mime_type: string; width: number; height: number; created_at: string;
}
export const media = {
  upload: (file: File) => { const fd = new FormData(); fd.append('file', file); return request<{ id: string; file_path: string }>('/admin/upload', { method: 'POST', body: fd }); },
  list: (params?: Record<string, number>) => request<{ media: { id: string; file_name: string; original_name: string; file_path: string; file_size: number; mime_type: string; created_at: string }[]; total: number; page: number; per_page: number }>('/admin/media' + qs(params)),
  delete: (id: string) => request(`/admin/media/${id}`, { method: 'DELETE' }),
};

// Tags
export const tags = {
  list: () => request<{ tags: { name: string; count: number }[] }>('/tags'),
};

// Settings
export const settings = {
  get: () => request<{ settings: any }>('/admin/settings'),
  update: (updates: Record<string, string>) => request<{ message: string }>('/admin/settings', { method: 'PUT', body: JSON.stringify(updates) }),
  public: () => request<{ site: any; theme: any; comment_enabled: boolean; social: any; custom_css: string; footer_text: string }>('/settings/public'),
};

// Tools
export const tools = {
  export: async () => {
    const url = `${BASE}/admin/export`;
    const headers: Record<string, string> = {};
    if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
    const res = await fetch(url, { headers, credentials: 'include' });
    if (!res.ok) throw new ApiError(res.status, 'EXPORT_ERROR', 'Export failed');
    const blob = await res.blob();
    const objUrl = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = objUrl;
    a.download = `mo-export-${new Date().toISOString().slice(0, 10)}.zip`;
    a.click();
    URL.revokeObjectURL(objUrl);
  },
  import: async (file: File) => {
    const fd = new FormData();
    fd.append('file', file);
    const url = `${BASE}/admin/import`;
    const headers: Record<string, string> = {};
    if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
    const res = await fetch(url, { method: 'POST', headers, credentials: 'include', body: fd });
    const data = await res.json();
    if (!res.ok) throw new ApiError(res.status, data.error?.code || 'IMPORT_ERROR', data.error?.message || 'Import failed');
    return data;
  },
  backup: () => request<{ message: string; file: string; location: string }>('/admin/backup', { method: 'POST' }),
  integrity: () => request<{ status: string; result: string }>('/admin/integrity'),
};

function qs(params?: Record<string, string | number>) {
  if (!params) return '';
  return '?' + new URLSearchParams(Object.entries(params).map(([k, v]) => [k, String(v)])).toString();
}

export const api = { auth, setup, posts, front, comments, pages, media, tags, settings, tools };
