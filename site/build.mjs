import { cp, mkdir } from 'node:fs/promises'

await mkdir('dist/site', { recursive: true })
for (const file of ['index.html', 'styles.css', 'main.js', 'favicon.svg']) {
  await cp('site/' + file, 'dist/site/' + file)
}
