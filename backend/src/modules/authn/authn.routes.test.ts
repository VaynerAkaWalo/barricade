import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { Elysia } from "elysia";
import { initalizeTables } from "../../infra/db.migrator";
import { authNRoutes } from "./authn.routes";
import { SessionStore } from "./session.store";

const TEST_EMAIL = "alice@test.com";
const TEST_PASSWORD = "my-secret-password";

function formBody(params: Record<string, string>): string {
	return new URLSearchParams(params).toString();
}

async function seedUser(db: Database): Promise<string> {
	const secretHash = await Bun.password.hash(TEST_PASSWORD);
	db.query(
		"INSERT INTO users (id, email, secret_hash, created_at, updated_at) VALUES ($id, $email, $secretHash, $now, $now)",
	).run({
		$id: "seed-user-id",
		$email: TEST_EMAIL,
		$secretHash: secretHash,
		$now: new Date().toISOString(),
	});
	return "seed-user-id";
}

function loginViaForm(
	app: ReturnType<typeof createApp>["app"],
	email: string,
	password: string,
): Promise<Response> {
	return app.handle(
		new Request("http://localhost/login", {
			method: "POST",
			headers: { "content-type": "application/x-www-form-urlencoded" },
			body: formBody({ email, password }),
		}),
	);
}

type TestApp = ReturnType<typeof createApp>;

function createApp() {
	const db = new Database(":memory:");
	initalizeTables(db);
	const app = new Elysia().use(authNRoutes(db));
	return { db, app };
}

describe("POST /login", () => {
	let app: TestApp["app"];
	let db: TestApp["db"];

	afterEach(() => {
		db?.close();
	});

	it("redirects to /dashboard and sets session cookie for valid credentials", async () => {
		({ db, app } = createApp());
		await seedUser(db);

		const res = await loginViaForm(app, TEST_EMAIL, TEST_PASSWORD);

		expect(res.status).toBe(302);
		expect(res.headers.get("location")).toBe("/dashboard");
		const setCookie = res.headers.get("set-cookie");
		expect(setCookie).toBeDefined();
		expect(setCookie).toContain("session=");
		expect(setCookie).toContain("HttpOnly");
	});

	it("returns 422 when email is missing", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/login", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ password: TEST_PASSWORD }),
			}),
		);

		expect(res.status).toBe(422);
	});

	it("returns 422 when password is missing", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/login", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ email: TEST_EMAIL }),
			}),
		);

		expect(res.status).toBe(422);
	});

	it("returns 401 with error message for invalid credentials", async () => {
		({ db, app } = createApp());
		await seedUser(db);

		const res = await loginViaForm(app, TEST_EMAIL, "wrong-password");

		expect(res.status).toBe(401);
		const text = await res.text();
		expect(text).toContain("Invalid email or password");
	});
});

describe("POST /logout", () => {
	let app: TestApp["app"];
	let db: TestApp["db"];

	afterEach(() => {
		db?.close();
	});

	it("redirects to /login and clears session cookie", async () => {
		({ db, app } = createApp());
		await seedUser(db);

		const loginRes = await loginViaForm(app, TEST_EMAIL, TEST_PASSWORD);
		const cookie = loginRes.headers.get("set-cookie") as string;

		const res = await app.handle(
			new Request("http://localhost/logout", {
				method: "POST",
				headers: { cookie },
			}),
		);

		expect(res.status).toBe(302);
		expect(res.headers.get("location")).toBe("/login");
		const setCookie = res.headers.get("set-cookie");
		expect(setCookie).toBeDefined();
		expect(setCookie).toContain("Max-Age=0");
	});

	it("redirects to /login without a session cookie", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/logout", {
				method: "POST",
			}),
		);

		expect(res.status).toBe(302);
		expect(res.headers.get("location")).toBe("/login");
	});
});

describe("GET /authenticate", () => {
	let app: TestApp["app"];
	let db: TestApp["db"];

	afterEach(() => {
		db?.close();
	});

	it("returns 200 with session info for valid cookie", async () => {
		({ db, app } = createApp());
		const userId = await seedUser(db);

		const loginRes = await loginViaForm(app, TEST_EMAIL, TEST_PASSWORD);
		const cookie = loginRes.headers.get("set-cookie") as string;

		const res = await app.handle(
			new Request("http://localhost/authenticate", {
				method: "GET",
				headers: { cookie },
			}),
		);

		expect(res.status).toBe(200);

		const body = (await res.json()) as Record<string, unknown>;
		expect(body.owner).toBe(userId);
		expect(body.expireAt).toBeDefined();
	});

	it("returns 401 for missing session cookie", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/authenticate", { method: "GET" }),
		);

		expect(res.status).toBe(401);
	});

	it("returns 401 for expired session", async () => {
		({ db, app } = createApp());
		await seedUser(db);

		const sessionStore = new SessionStore(db);
		await sessionStore.createSession({
			id: "expired-session",
			owner: "seed-user-id",
			fingerPrint: Bun.hash.wyhash("test-agent"),
			createdAt: new Date("2020-01-01T00:00:00Z"),
			expireAt: new Date("2020-01-01T00:05:00Z"),
		});

		const res = await app.handle(
			new Request("http://localhost/authenticate", {
				method: "GET",
				headers: {
					cookie: "session=expired-session",
					"user-agent": "test-agent",
				},
			}),
		);

		expect(res.status).toBe(401);
	});

	it("returns 401 for fingerprint mismatch", async () => {
		({ db, app } = createApp());
		await seedUser(db);

		const sessionStore = new SessionStore(db);
		await sessionStore.createSession({
			id: "valid-session",
			owner: "seed-user-id",
			fingerPrint: Bun.hash.wyhash("original-agent"),
			createdAt: new Date(),
			expireAt: new Date(Date.now() + 5 * 60 * 1000),
		});

		const res = await app.handle(
			new Request("http://localhost/authenticate", {
				method: "GET",
				headers: {
					cookie: "session=valid-session",
					"user-agent": "different-agent",
				},
			}),
		);

		expect(res.status).toBe(401);
	});
});
