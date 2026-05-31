import { Elysia } from 'elysia'
import { staticPlugin } from '@elysiajs/static'

const app = new Elysia()
  .use(staticPlugin({
    assets: "public/assets",
    prefix: "/assets"
  }
  ))

  .get("/health", () => "ok")

  .get('*', () => Bun.file('./public/index.html'))
  .listen(3000)

console.log(`Server started on port ${app.server?.port}`)
