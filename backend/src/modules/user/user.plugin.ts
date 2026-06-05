import type { Database } from "bun:sqlite";
import Elysia from "elysia";
import { UserManagementService } from "./user.service";
import { UserManagementStore } from "./user.store";

export const userPlugin = (db: Database) =>
	new Elysia({ name: "module.user" }).decorate(() => {
		const userStore = new UserManagementStore(db);
		const userService = new UserManagementService(userStore);

		return {
			userStore,
			userService,
		};
	});
