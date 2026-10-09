export class ApiError extends Error {
  readonly status: number;
  readonly fields: Record<string, string>;

  constructor(status: number, message: string, fields: Record<string, string> = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = fields;
  }
}

interface Problem {
  title?: string;
  detail?: string;
  errors?: Record<string, string>;
}

export interface ApiResponse<T> {
  data: T;
  headers: Headers;
  status: number;
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<ApiResponse<T>> {
  const headers = new Headers(init.headers);
  if (init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  let response: Response;
  try {
    response = await fetch(path, { ...init, headers });
  } catch {
    throw new ApiError(0, "Could not reach the server. Check that the backend is running.");
  }

  if (!response.ok) {
    const problem = (await response.json().catch(() => ({}))) as Problem;
    throw new ApiError(response.status, problem.detail ?? problem.title ?? response.statusText, problem.errors);
  }

  const data = response.status === 204 ? undefined : await response.json();
  return { data: data as T, headers: response.headers, status: response.status };
}
