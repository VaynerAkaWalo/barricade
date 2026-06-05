export class ApiError extends Error {
	status: number;
	constructor(message: string, status: number) {
		super(message);
		this.status = status;
	}
}

interface Credentials {
	email: string;
	secret: string;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(path, {
		credentials: "include",
		...options,
		headers: { "Content-Type": "application/json", ...options?.headers },
	});

	if (!res.ok) {
		let message = "An error occurred";
		try {
			const body = await res.json();
			message = body.message || body.error || JSON.stringify(body);
		} catch {
			// response body is not JSON, ignore
		}
		throw new ApiError(message, res.status);
	}

	if (res.status === 204) return undefined as T;

	const text = await res.text();
	if (!text) return undefined as T;
	return JSON.parse(text);
}

export function login(params: Credentials): Promise<void> {
	return request("/login", {
		method: "POST",
		body: JSON.stringify(params),
	});
}

export function register(params: Credentials): Promise<void> {
	return request("/users", {
		method: "POST",
		body: JSON.stringify(params),
	});
}

export function authenticate(): Promise<void> {
	return request("/authenticate");
}

export function logout(): Promise<void> {
	return request("/logout", { method: "POST" });
}
