import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { Elysia } from "elysia";
import { initalizeTables } from "../../infra/db.migrator";
import { userRoutes } from "./user.routes";

type TestApp = ReturnType<typeof createApp>;

function formBody(params: Record<string, string>): string {
	return new URLSearchParams(params).toString();
}

function createApp() {
	const db = new Database(":memory:");
	initalizeTables(db);
	const app = new Elysia().use(userRoutes(db));
	return { db, app };
}

const TEST_EMAIL = "alice@test.com";
const TEST_PASSWORD = "pass123";

describe("POST /register", () => {
	let app: TestApp["app"];
	let db: TestApp["db"];

	afterEach(() => {
		db.close();
	});

	it("creates a user and redirects to /login", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/register", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ email: TEST_EMAIL, password: TEST_PASSWORD }),
			}),
		);

		expect(res.status).toBe(302);
		expect(res.headers.get("location")).toBe("/login");
	});

	it("returns 409 for duplicate email", async () => {
		({ db, app } = createApp());

		await app.handle(
			new Request("http://localhost/register", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ email: TEST_EMAIL, password: TEST_PASSWORD }),
			}),
		);

		const res = await app.handle(
			new Request("http://localhost/register", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ email: TEST_EMAIL, password: "pass456" }),
			}),
		);

		expect(res.status).toBe(409);
		const text = await res.text();
		expect(text).toContain("An account with this email already exists");
	});

	it("returns 422 when email is missing", async () => {
		({ db, app } = createApp());

		const res = await app.handle(
			new Request("http://localhost/register", {
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
			new Request("http://localhost/register", {
				method: "POST",
				headers: { "content-type": "application/x-www-form-urlencoded" },
				body: formBody({ email: TEST_EMAIL }),
			}),
		);

		expect(res.status).toBe(422);
	});
});
