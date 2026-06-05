import { Database } from "bun:sqlite";
import { staticPlugin } from "@elysiajs/static";
import { Elysia } from "elysia";
import { initalizeTables } from "./infra/db.migrator";
import { authNRoutes } from "./modules/authn/authn.routes";
import { userRoutes } from "./modules/user/user.routes";

const db: Database = new Database(":memory:");

initalizeTables(db);

const app = new Elysia()
	.use(
		staticPlugin({
			assets: "public/assets",
			prefix: "/assets",
		}),
	)

	.use(userRoutes(db))

	.use(authNRoutes(db))

	.get("/health", () => "ok")

	.get("*", () => Bun.file("./public/index.html"))
	.listen(3000);

console.log(`Server started on port ${app.server?.port}`);
