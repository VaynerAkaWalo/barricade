import type { Database } from "bun:sqlite";
import { Elysia, t } from "elysia";
import { userPlugin } from "./user.plugin";

const UserDTO = t.Object({
	id: t.String(),
	email: t.String(),
});

export const userRoutes = (db: Database) =>
	new Elysia({ prefix: "/users" }).use(userPlugin(db)).post(
		"/",
		async ({ body, userService }) => {
			const user = await userService.createUser({
				email: body.email,
				secret: body.secret,
			});

			return user;
		},
		{
			body: t.Object({
				email: t.String(),
				secret: t.String(),
			}),
			response: UserDTO,
		},
	);
