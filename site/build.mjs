import { cp, mkdir } from 'node:fs/promises'

await mkdir('dist/site', { recursive: true })
for (const file of ['index.html', 'styles.css', 'main.js', 'favicon.svg', 'testimonials.json']) {
  await cp('site/' + file, 'dist/site/' + file)
}
await cp('install.sh', 'dist/site/install.sh')
