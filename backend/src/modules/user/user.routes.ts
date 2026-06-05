import type { Database } from "bun:sqlite";
import { Elysia, t } from "elysia";
import { DuplicateEmailError } from "./user.errors";
import { userPlugin } from "./user.plugin";

const UserDTO = t.Object({
	id: t.String(),
	email: t.String(),
});

const ErrorResponse = t.Object({ error: t.String() });

export const userRoutes = (db: Database) =>
	new Elysia({ prefix: "/users" })
		.use(userPlugin(db))
		.onError(({ error, set }) => {
			if (error instanceof DuplicateEmailError) {
				set.status = 409;
				return { error: error.message };
			}
		})
		.post(
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
				response: {
					200: UserDTO,
					409: ErrorResponse,
				},
			},
		);
