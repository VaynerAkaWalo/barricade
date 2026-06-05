import type { Database } from "bun:sqlite";
import Elysia from "elysia";
import { userPlugin } from "../user/user.plugin";
import { AuthenticationService } from "./authn.service";
import { SessionStore } from "./session.store";

export const authenticationPlugin = (db: Database) =>
	new Elysia({ name: "module.authn" }).use(
		userPlugin(db).decorate(({ userStore }) => {
			const sessionStore = new SessionStore(db);
			const authenticationService = new AuthenticationService(
				sessionStore,
				userStore,
			);

			return {
				sessionStore,
				authenticationService,
			};
		}),
	);
