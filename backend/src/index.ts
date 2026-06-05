import { Database } from "bun:sqlite";
import { Elysia } from "elysia";
import { initalizeTables } from "./infra/db.migrator";
import { authNRoutes } from "./modules/authn/authn.routes";
import { userRoutes } from "./modules/user/user.routes";

const db: Database = new Database(":memory:");

initalizeTables(db);

const app = new Elysia()
	.get("/styles.css", () => Bun.file("./public/styles.css"))

	.use(authNRoutes(db))

	.use(userRoutes(db))

	.get("/health", () => "ok")

	.get("*", ({ set }) => {
		set.status = 302;
		set.headers.location = "/login";
		return "";
	})
	.listen(3000);

console.log(`Server started on port ${app.server?.port}`);
