import { randomUUIDv7 } from "bun";
import type { UserManagementStore } from "../user/user.store";
import {
	AccountNotFoundError,
	FingerPrintMismach,
	SessionExpiredError,
	WrongPasswordError,
} from "./authn.errors";
import type { Session } from "./seession.entity";
import type { SessionStore } from "./session.store";

export interface LoginParams {
	email: string;
	secret: string;
	rawFingerPrint: string;
}

export class AuthenticationService {
	private readonly sessionStore: SessionStore;
	private readonly userStore: UserManagementStore;

	constructor(sessionStore: SessionStore, userStore: UserManagementStore) {
		this.sessionStore = sessionStore;
		this.userStore = userStore;
	}

	async createSesssion({
		email,
		secret,
		rawFingerPrint,
	}: LoginParams): Promise<Session> {
		const user = await this.userStore.getUserByEmail(email);
		if (user == null) {
			throw new AccountNotFoundError(
				"Account with provided email does not exists",
			);
		}

		if (!(await Bun.password.verify(secret, user.secretHash))) {
			throw new WrongPasswordError();
		}

		const createdAt = new Date();
		const expireAt = new Date(createdAt.getTime() + 5 * 60 * 1000);

		const newSession: Session = {
			id: randomUUIDv7(),
			owner: user.id,
			fingerPrint: Bun.hash.wyhash(rawFingerPrint),
			createdAt: createdAt,
			expireAt: expireAt,
		};

		await this.sessionStore.createSession(newSession);

		return newSession;
	}

	async authenticate(id: string, rawFingerPrint: string): Promise<Session> {
		const session = await this.sessionStore.getSession(id);
		if (session == null || session.expireAt < new Date()) {
			throw new SessionExpiredError();
		}

		if (
			Number(session.fingerPrint) !== Number(Bun.hash.wyhash(rawFingerPrint))
		) {
			throw new FingerPrintMismach();
		}

		return session;
	}
}
