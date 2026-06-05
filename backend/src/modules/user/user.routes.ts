import type { Database } from "bun:sqlite";
import { Elysia, t } from "elysia";
import { layout } from "../../templates/layout";
import { registerPage } from "../../templates/pages/register";
import { DuplicateEmailError } from "./user.errors";
import { userPlugin } from "./user.plugin";

function getRegisterErrorMessage(error: unknown): string {
	if (error instanceof DuplicateEmailError) {
		return "An account with this email already exists";
	}
	return "An unexpected error occurred";
}

export const userRoutes = (db: Database) =>
	new Elysia()
		.use(userPlugin(db))
		.onError(({ error, set }) => {
			if (error instanceof DuplicateEmailError) {
				set.status = 409;
				return { error: error.message };
			}
		})
		.get("/register", ({ set }) => {
			set.headers["content-type"] = "text/html; charset=utf-8";
			return layout(registerPage({}), { title: "Create Account" });
		})
		.post(
			"/register",
			async ({ body, set, userService }) => {
				try {
					await userService.createUser({
						email: body.email,
						secret: body.password,
					});

					set.status = 302;
					set.headers.location = "/login";
					return "";
				} catch (error) {
					const message = getRegisterErrorMessage(error);
					set.headers["content-type"] = "text/html; charset=utf-8";
					set.status = 409;
					return layout(registerPage({ email: body.email, error: message }), {
						title: "Create Account",
					});
				}
			},
			{
				body: t.Object({
					email: t.String(),
					password: t.String(),
				}),
			},
		);
