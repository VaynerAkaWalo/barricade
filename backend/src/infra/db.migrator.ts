import type { Database } from "bun:sqlite";

export interface Table {
	name: string;
	script: string;
}

const tables: Table[] = [
	{
		name: "users",
		script: `
    CREATE TABLE users (
      id TEXT PRIMARY KEY,
      email TEXT UNIQUE,
      secret_hash TEXT,
      created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
      updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
    )
    `,
	},
	{
		name: "sessions",
		script: `
     CREATE TABLE sessions (
       id TEXT PRIMARY KEY,
       owner_id TEXT KEY,
       finger_print INTEGER,
       created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
       expire_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
     )
     `,
	},
];

export const initalizeTables = (db: Database) => {
	console.log(
		`Database migration attempt for tables [${tables.map((table) => table.name).join(",")}]`,
	);
	tables.forEach((table) => {
		try {
			db.query(table.script).run();
		} catch (error) {
			console.log(`Error while initalizing table ${table.name}`);
			throw error;
		}
	});
	console.log("Migration completed sucessfully");
};
