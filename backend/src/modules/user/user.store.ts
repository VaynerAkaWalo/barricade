import type { Database } from "bun:sqlite";
import type { User } from "./user.entity";

export class UserManagementStore {
	private readonly db: Database;

	constructor(db: Database) {
		this.db = db;
	}

	async CreateUser(user: User): Promise<User> {
		const insertQuery = this.db.query(
			"INSERT INTO users (id, email, secret_hash) VALUES ($id, $email, $secretHash)",
		);

		insertQuery.run({
			$id: user.id,
			$email: user.email,
			$secretHash: user.secretHash,
		});

		return user;
	}

	async GetUser(id: string): Promise<User | null> {
		const row = this.db
			.query(
				"SELECT id, email, secret_hash, created_at, updated_at FROM users WHERE id = $id",
			)
			.get({ $id: id }) as Record<string, unknown> | undefined;

		if (!row) return null;

		return {
			id: row.id as string,
			email: row.email as string,
			secretHash: row.secret_hash as string,
			createdAt: new Date(row.created_at as string),
			updatedAt: new Date(row.updated_at as string),
		};
	}
}
