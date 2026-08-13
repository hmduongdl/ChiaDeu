const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL || "/api").replace(/\/$/, "");

type UnauthorizedHandler = () => void;

let refreshPromise: Promise<boolean> | null = null;
let unauthorizedHandler: UnauthorizedHandler = () => undefined;

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export function setUnauthorizedHandler(handler: UnauthorizedHandler) {
  unauthorizedHandler = handler;
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
  options: { skipAuthRefresh?: boolean } = {},
): Promise<T> {
  const response = await request(path, init);

  if (response.status === 401 && !options.skipAuthRefresh) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      return parseResponse<T>(await request(path, init));
    }
    unauthorizedHandler();
  }

  return parseResponse<T>(response);
}

async function request(path: string, init: RequestInit) {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  return fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
    // Tokens are HttpOnly cookies; credentials is required for every authenticated call.
    credentials: "include",
  });
}

async function refreshAccessToken() {
  if (!refreshPromise) {
    refreshPromise = request("/auth/refresh", { method: "POST" })
      .then((response) => response.ok)
      .catch(() => false)
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

async function parseResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    if (response.status === 204) {
      return undefined as T;
    }
    return response.json() as Promise<T>;
  }

  let message = "Request failed";
  try {
    const body = (await response.json()) as { error?: string };
    message = body.error || message;
  } catch {
    // Keep a stable fallback for non-JSON upstream errors.
  }
  throw new ApiError(response.status, message);
}
