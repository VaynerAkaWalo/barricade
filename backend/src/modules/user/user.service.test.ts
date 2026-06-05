import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { initalizeTables } from "../../infra/db.migrator";
import { DuplicateEmailError } from "./user.errors";
import { UserManagementService } from "./user.service";
import { UserManagementStore } from "./user.store";

function createDb(): Database {
	const db = new Database(":memory:");
	initalizeTables(db);
	return db;
}

describe("UserManagementService", () => {
	let db: Database;
	let service: UserManagementService;

	afterEach(() => {
		db.close();
	});

	it("creates a user with all required fields", async () => {
		db = createDb();
		service = new UserManagementService(new UserManagementStore(db));

		const user = await service.createUser({
			email: "bob@test.com",
			secret: "my-password",
		});

		expect(user.id).toBeDefined();
		expect(typeof user.id).toBe("string");
		expect(user.email).toBe("bob@test.com");
		expect(user.createdAt).toBeInstanceOf(Date);
		expect(user.updatedAt).toBeInstanceOf(Date);
	});

	it("hashes the provided secret", async () => {
		db = createDb();
		service = new UserManagementService(new UserManagementStore(db));

		const secret = "super-secret-123";
		const user = await service.createUser({ email: "bob@test.com", secret });

		expect(user.secretHash).toBeDefined();
		expect(user.secretHash).not.toBe(secret);
	});

	it("throws DuplicateEmailError for duplicate email", async () => {
		db = createDb();
		service = new UserManagementService(new UserManagementStore(db));

		await service.createUser({ email: "bob@test.com", secret: "my-password" });

		expect(
			service.createUser({ email: "bob@test.com", secret: "other-password" }),
		).rejects.toThrow(DuplicateEmailError);
	});

	it("persists the user in the database", async () => {
		db = createDb();
		service = new UserManagementService(new UserManagementStore(db));

		const user = await service.createUser({
			email: "bob@test.com",
			secret: "my-password",
		});

		const store = new UserManagementStore(db);
		const found = await store.getUser(user.id);
		expect(found).toBeDefined();
		expect(found?.id).toBe(user.id);
		expect(found?.email).toBe(user.email);
		expect(found?.secretHash).toBe(user.secretHash);
	});
});
