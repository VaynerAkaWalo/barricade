import type { Database } from "bun:sqlite";
import { Elysia, t } from "elysia";
import { layout } from "../../templates/layout";
import { dashboardPage } from "../../templates/pages/dashboard";
import { loginPage } from "../../templates/pages/login";
import {
	AccountNotFoundError,
	FingerPrintMismach,
	SessionExpiredError,
	WrongPasswordError,
} from "./authn.errors";
import { authenticationPlugin } from "./authn.plugin";

const AuthenticationDTO = t.Object({
	owner: t.String(),
	expireAt: t.String(),
});

function getFormErrorMessage(error: unknown): string {
	if (
		error instanceof AccountNotFoundError ||
		error instanceof WrongPasswordError
	) {
		return "Invalid email or password";
	}
	return "An unexpected error occurred";
}

export const authNRoutes = (db: Database) =>
	new Elysia()
		.use(authenticationPlugin(db))
		.onError(({ error, set }) => {
			if (
				error instanceof AccountNotFoundError ||
				error instanceof WrongPasswordError
			) {
				set.status = 401;
				return { error: "Invalid credentials" };
			}
			if (error instanceof SessionExpiredError) {
				set.status = 401;
				return { error: "Session expired" };
			}
			if (error instanceof FingerPrintMismach) {
				set.status = 401;
				return { error: "Unauthorized" };
			}
		})
		.get(
			"/login",
			async ({ cookie: { session }, headers, authenticationService, set }) => {
				if (session.value) {
					try {
						await authenticationService.authenticate(
							session.value,
							headers["User-Agent"] ?? "",
						);
						set.status = 302;
						set.headers.location = "/dashboard";
						return "";
					} catch {}
				}

				set.headers["content-type"] = "text/html; charset=utf-8";
				return layout(loginPage({}), { title: "Sign In" });
			},
			{
				cookie: t.Cookie({
					session: t.Optional(t.String()),
				}),
			},
		)
		.get(
			"/dashboard",
			async ({ cookie: { session }, headers, authenticationService, set }) => {
				if (!session.value) {
					set.status = 302;
					set.headers.location = "/login";
					return "";
				}
				try {
					const userSession = await authenticationService.authenticate(
						session.value,
						headers["User-Agent"] ?? "",
					);

					set.headers["content-type"] = "text/html; charset=utf-8";
					return layout(dashboardPage({ ownerId: userSession.owner }), {
						title: "Dashboard",
						isAuthenticated: true,
					});
				} catch {
					set.status = 302;
					set.headers.location = "/login";
					return "";
				}
			},
			{
				cookie: t.Cookie({
					session: t.Optional(t.String()),
				}),
			},
		)
		.post(
			"/login",
			async ({
				body,
				cookie: { session },
				headers,
				set,
				authenticationService,
			}) => {
				try {
					const createdSession = await authenticationService.createSesssion({
						email: body.email,
						secret: body.password,
						rawFingerPrint: headers["User-Agent"] ?? "",
					});

					session.value = createdSession.id;
					session.httpOnly = true;
					session.maxAge = 300;

					set.status = 302;
					set.headers.location = "/dashboard";
					return "";
				} catch (error) {
					const message = getFormErrorMessage(error);
					set.headers["content-type"] = "text/html; charset=utf-8";
					set.status = 401;
					return layout(loginPage({ email: body.email, error: message }), {
						title: "Sign In",
					});
				}
			},
			{
				body: t.Object({
					email: t.String(),
					password: t.String(),
				}),
				cookie: t.Cookie({
					session: t.Optional(t.String()),
				}),
			},
		)
		.post(
			"/logout",
			({ cookie: { session }, set }) => {
				session.value = "";
				session.maxAge = 0;

				set.status = 302;
				set.headers.location = "/login";
				return "";
			},
			{
				cookie: t.Cookie({
					session: t.Optional(t.String()),
				}),
			},
		)
		.get(
			"/authenticate",
			async ({ cookie: { session }, headers, authenticationService }) => {
				if (!session.value) {
					throw new SessionExpiredError();
				}

				const userSession = await authenticationService.authenticate(
					session.value,
					headers["User-Agent"] ?? "",
				);

				return {
					owner: userSession.owner,
					expireAt: userSession.expireAt.toISOString(),
				};
			},
			{
				cookie: t.Cookie({
					session: t.Optional(t.String()),
				}),
				response: AuthenticationDTO,
			},
		);
