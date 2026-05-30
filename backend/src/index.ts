import { Elysia } from 'elysia'

const app = new Elysia()
  .get("/health", () => "ok")
  .listen(3000)

console.log(`Server started on port ${app.server?.port}`)
