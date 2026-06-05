import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { initalizeTables } from "../../infra/db.migrator";
import type { User } from "./user.entity";
import { UserManagementStore } from "./user.store";

function createDb(): Database {
	const db = new Database(":memory:");
	initalizeTables(db);
	return db;
}

function makeUser(overrides: Partial<User> = {}): User {
	return {
		id: "test-id-1",
		email: "alice@test.com",
		secretHash: "hashed-secret",
		createdAt: new Date("2025-01-01T00:00:00Z"),
		updatedAt: new Date("2025-01-01T00:00:00Z"),
		...overrides,
	};
}

describe("UserManagementStore", () => {
	let db: Database;
	let store: UserManagementStore;

	afterEach(() => {
		db.close();
	});

	it("creates a user and returns it", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		const input = makeUser();
		const result = await store.createUser(input);

		expect(result).toEqual(input);
	});

	it("getUser returns the user after creation", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		const input = makeUser();
		await store.createUser(input);

		const found = await store.getUser(input.id);
		expect(found).toBeDefined();
		expect(found?.id).toBe(input.id);
		expect(found?.email).toBe(input.email);
		expect(found?.secretHash).toBe(input.secretHash);
	});

	it("getUser returns null for non-existent user", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		const found = await store.getUser("non-existent-id");
		expect(found).toBeNull();
	});

	it("getUserByEmail returns the user after creation", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		const input = makeUser();
		await store.createUser(input);

		const found = await store.getUserByEmail(input.email);
		expect(found).toBeDefined();
		expect(found?.id).toBe(input.id);
		expect(found?.email).toBe(input.email);
		expect(found?.secretHash).toBe(input.secretHash);
	});

	it("getUserByEmail returns null for non-existent email", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		const found = await store.getUserByEmail("doesnotexist@test.com");
		expect(found).toBeNull();
	});

	it("throws when inserting a duplicate email", async () => {
		db = createDb();
		store = new UserManagementStore(db);

		await store.createUser(makeUser());

		expect(
			store.createUser(
				makeUser({ id: "different-id", email: "alice@test.com" }),
			),
		).rejects.toThrow();
	});
});
