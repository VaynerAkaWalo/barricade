import type { Database } from "bun:sqlite";
import { Elysia, t } from "elysia";
import { SessionExpiredError } from "./authn.errors";
import { authenticationPlugin } from "./authn.plugin";

const AuthenticationDTO = t.Object({
	owner: t.String(),
	expireAt: t.String(),
});

export const authNRoutes = (db: Database) =>
	new Elysia()
		.use(authenticationPlugin(db))
		.post(
			"/login",
			async ({ body, cookie: { session }, headers, authenticationService }) => {
				const createdSession = await authenticationService.createSesssion({
					email: body.email,
					secret: body.secret,
					rawFingerPrint: headers["User-Agent"] ?? "",
				});

				session.value = createdSession.id;
				session.httpOnly = true;
				session.maxAge = 300;
			},
			{
				body: t.Object({
					email: t.String(),
					secret: t.String(),
				}),
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
