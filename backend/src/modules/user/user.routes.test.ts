import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { Elysia } from "elysia";
import { initalizeTables } from "../../infra/db.migrator";
import { userRoutes } from "./user.routes";

type TestApp = ReturnType<typeof createApp>;

function createApp() {
	const db = new Database(":memory:");
	initalizeTables(db);
	const app = new Elysia().use(userRoutes(db));
	return { db, app };
}

describe("POST /users", () => {
	let app: TestApp["app"];
	let db: TestApp["db"];

	afterEach(() => {
		db.close();
	});

	it("creates a user and returns it", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/users", {
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ email: "alice@test.com", secret: "pass123" }),
			}),
		);

		expect(res.status).toBe(200);

		const body = (await res.json()) as Record<string, unknown>;
		expect(body.id).toBeDefined();
		expect(body.email).toBe("alice@test.com");
	});

	it("returns 422 when email is missing", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/users", {
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ secret: "pass123" }),
			}),
		);

		expect(res.status).toBe(422);
	});

	it("returns 422 when secret is missing", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/users", {
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ email: "alice@test.com" }),
			}),
		);

		expect(res.status).toBe(422);
	});

	it("returns 422 for non-JSON body", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/users", {
				method: "POST",
				headers: { "content-type": "text/plain" },
				body: "not json",
			}),
		);

		expect(res.status).toBe(422);
	});
});

describe("GET /users", () => {
	it("returns 404", async () => {
		const { app } = createApp();

		const res = await app.handle(
			new Request("http://localhost/users", { method: "GET" }),
		);

		expect(res.status).toBe(404);
	});
});
