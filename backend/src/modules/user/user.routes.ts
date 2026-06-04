import { Elysia, t } from "elysia";
import type { UserManagementService } from "./user.service";

const UserDTO = t.Object({
	id: t.String(),
	email: t.String(),
});

export const userRoutes = (userService: UserManagementService) =>
	new Elysia({ prefix: "/users" }).post(
		"/",
		async ({ body }) => {
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
