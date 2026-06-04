import type { Database } from "bun:sqlite";
import { randomUUIDv7 } from "bun";
import type { User } from "./user.entity";
import { UserManagementStore as UserStore } from "./user.store";

export interface CreateUserParams {
	email: string;
	secret: string;
}

export class UserManagementService {
	private readonly store: UserStore;

	constructor(db: Database) {
		this.store = new UserStore(db);
	}

	async createUser({ email, secret }: CreateUserParams): Promise<User> {
		const secretHash = await Bun.password.hash(secret);

		const user: User = {
			id: randomUUIDv7(),
			email,
			secretHash,
			createdAt: new Date(),
			updatedAt: new Date(),
		};

		console.log(`Creating user ${user.id} with email ${email}`);

		return this.store.CreateUser(user);
	}
}
