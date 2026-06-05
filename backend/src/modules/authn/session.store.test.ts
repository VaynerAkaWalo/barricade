import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { initalizeTables } from "../../infra/db.migrator";
import type { Session } from "./seession.entity";
import { SessionStore } from "./session.store";

function createDb(): Database {
	const db = new Database(":memory:");
	initalizeTables(db);
	return db;
}

function makeSession(overrides: Partial<Session> = {}): Session {
	return {
		id: "test-session-1",
		owner: "test-user-1",
		fingerPrint: 12345n,
		createdAt: new Date("2025-01-01T00:00:00Z"),
		expireAt: new Date("2025-01-01T00:05:00Z"),
		...overrides,
	};
}

describe("SessionStore", () => {
	let db: Database;
	let store: SessionStore;

	afterEach(() => {
		db.close();
	});

	it("creates a session and returns it", async () => {
		db = createDb();
		store = new SessionStore(db);

		const input = makeSession();
		const result = await store.createSession(input);

		expect(result).toEqual(input);
	});

	it("getSession returns the session after creation", async () => {
		db = createDb();
		store = new SessionStore(db);

		const input = makeSession();
		await store.createSession(input);

		const found = await store.getSession(input.id);
		expect(found).toBeDefined();
		expect(found?.id).toBe(input.id);
		expect(found?.owner).toBe(input.owner);
		expect(Number(found?.fingerPrint)).toBe(Number(input.fingerPrint));
	});

	it("getSession returns null for non-existent session", async () => {
		db = createDb();
		store = new SessionStore(db);

		const found = await store.getSession("non-existent-id");
		expect(found).toBeNull();
	});
});
