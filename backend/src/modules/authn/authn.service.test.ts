import { Database } from "bun:sqlite";
import { afterEach, describe, expect, it } from "bun:test";
import { initalizeTables } from "../../infra/db.migrator";
import { UserManagementStore } from "../user/user.store";
import {
	AccountNotFoundError,
	FingerPrintMismach,
	SessionExpiredError,
	WrongPasswordError,
} from "./authn.errors";
import { AuthenticationService } from "./authn.service";
import { SessionStore } from "./session.store";

function createDb(): Database {
	const db = new Database(":memory:");
	initalizeTables(db);
	return db;
}

const TEST_EMAIL = "alice@test.com";
const TEST_PASSWORD = "my-secret-password";

async function seedUser(userStore: UserManagementStore): Promise<string> {
	const secretHash = await Bun.password.hash(TEST_PASSWORD);
	const user = {
		id: "seed-user-id",
		email: TEST_EMAIL,
		secretHash,
		createdAt: new Date(),
		updatedAt: new Date(),
	};
	await userStore.createUser(user);
	return user.id;
}

describe("createSesssion", () => {
	let db: Database;
	let sessionStore: SessionStore;
	let userStore: UserManagementStore;
	let service: AuthenticationService;

	afterEach(() => {
		db.close();
	});

	it("creates a session with valid credentials", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);
		await seedUser(userStore);

		const session = await service.createSesssion({
			email: TEST_EMAIL,
			secret: TEST_PASSWORD,
			rawFingerPrint: "test-agent",
		});

		expect(session).toBeDefined();
		expect(session.id).toBeDefined();
		expect(session.owner).toBe("seed-user-id");
		expect(session.fingerPrint).toBe(Bun.hash.wyhash("test-agent"));
		expect(session.createdAt).toBeInstanceOf(Date);
		expect(session.expireAt).toBeInstanceOf(Date);
		expect(session.expireAt.getTime() - session.createdAt.getTime()).toBe(
			5 * 60 * 1000,
		);
	});

	it("throws AccountNotFoundError for non-existent email", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);

		expect(
			service.createSesssion({
				email: "doesnotexist@test.com",
				secret: TEST_PASSWORD,
				rawFingerPrint: "test-agent",
			}),
		).rejects.toThrow(AccountNotFoundError);
	});

	it("throws WrongPasswordError for wrong password", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);
		await seedUser(userStore);

		expect(
			service.createSesssion({
				email: TEST_EMAIL,
				secret: "wrong-password",
				rawFingerPrint: "test-agent",
			}),
		).rejects.toThrow(WrongPasswordError);
	});

	it("persists the session in the database", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);
		await seedUser(userStore);

		const session = await service.createSesssion({
			email: TEST_EMAIL,
			secret: TEST_PASSWORD,
			rawFingerPrint: "test-agent",
		});

		const found = await sessionStore.getSession(session.id);
		expect(found).toBeDefined();
		expect(found?.id).toBe(session.id);
	});
});

describe("authenticate", () => {
	let db: Database;
	let sessionStore: SessionStore;
	let userStore: UserManagementStore;
	let service: AuthenticationService;

	afterEach(() => {
		db.close();
	});

	it("returns session for valid id and matching fingerprint", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);
		await seedUser(userStore);

		const session = await service.createSesssion({
			email: TEST_EMAIL,
			secret: TEST_PASSWORD,
			rawFingerPrint: "test-agent",
		});

		const result = await service.authenticate(session.id, "test-agent");
		expect(result).toBeDefined();
		expect(result.id).toBe(session.id);
	});

	it("throws SessionExpiredError for expired session", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);

		const expiredSession = {
			id: "expired-id",
			owner: "some-user",
			fingerPrint: Bun.hash.wyhash("test-agent"),
			createdAt: new Date("2020-01-01T00:00:00Z"),
			expireAt: new Date("2020-01-01T00:05:00Z"),
		};
		await sessionStore.createSession(expiredSession);

		expect(
			service.authenticate(expiredSession.id, "test-agent"),
		).rejects.toThrow(SessionExpiredError);
	});

	it("throws SessionExpiredError for non-existent session", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);

		expect(
			service.authenticate("non-existent-id", "test-agent"),
		).rejects.toThrow(SessionExpiredError);
	});

	it("throws FingerPrintMismach for mismatched fingerprint", async () => {
		db = createDb();
		userStore = new UserManagementStore(db);
		sessionStore = new SessionStore(db);
		service = new AuthenticationService(sessionStore, userStore);
		await seedUser(userStore);

		const session = await service.createSesssion({
			email: TEST_EMAIL,
			secret: TEST_PASSWORD,
			rawFingerPrint: "original-agent",
		});

		expect(
			service.authenticate(session.id, "different-agent"),
		).rejects.toThrow(FingerPrintMismach);
	});
});
