import { Database } from "bun:sqlite";
import { staticPlugin } from "@elysiajs/static";
import { Elysia } from "elysia";
import { initalizeTables } from "./infra/db.migrator";
import { userRoutes } from "./modules/user/user.routes";
import { UserManagementService } from "./modules/user/user.service";

const db: Database = new Database(":memory:");

initalizeTables(db);

const userService = new UserManagementService(db);

const app = new Elysia()
	.use(
		staticPlugin({
			assets: "public/assets",
			prefix: "/assets",
		}),
	)

	.use(userRoutes(userService))

	.get("/health", () => "ok")

	.get("*", () => Bun.file("./public/index.html"))
	.listen(3000);

console.log(`Server started on port ${app.server?.port}`);
