import type { Database } from "bun:sqlite";
import type { Session } from "./seession.entity";

export class SessionStore {
	private readonly db: Database;

	constructor(db: Database) {
		this.db = db;
	}

	async createSession(session: Session): Promise<Session> {
		const insertQuery = this.db.query(
			"INSERT INTO sessions (id, owner_id, finger_print, expire_at) VALUES ($id, $ownerId, $fingerPrint, $expireAt)",
		);

		insertQuery.run({
			$id: session.id,
			$ownerId: session.owner,
			$fingerPrint: session.fingerPrint,
			$expireAt: session.expireAt.toISOString(),
		});

		return session;
	}

	async getSession(id: string): Promise<Session | null> {
		const row = this.db
			.query(
				"SELECT id, owner_id, finger_print, expire_at, created_at FROM sessions WHERE id = $id",
			)
			.get({ $id: id }) as Record<string, unknown> | undefined;

		if (!row) return null;

		return {
			id: row.id as string,
			owner: row.owner_id as string,
			fingerPrint: row.finger_print as bigint,
			createdAt: new Date(row.created_at as string),
			expireAt: new Date(row.expire_at as string),
		};
	}
}
