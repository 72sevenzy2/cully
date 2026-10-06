const menuButton = document.querySelector('.menu-toggle')
const nav = document.querySelector('.site-nav')

menuButton?.addEventListener('click', () => {
  const open = menuButton.getAttribute('aria-expanded') !== 'true'
  menuButton.setAttribute('aria-expanded', String(open))
  menuButton.setAttribute('aria-label', open ? 'Close navigation' : 'Open navigation')
  nav.classList.toggle('is-open', open)
})

nav?.querySelectorAll('a').forEach(link => link.addEventListener('click', () => {
  menuButton.setAttribute('aria-expanded', 'false')
  menuButton.setAttribute('aria-label', 'Open navigation')
  nav.classList.remove('is-open')
}))

document.querySelectorAll('.copy-button').forEach(button => button.addEventListener('click', async () => {
  const code = button.parentElement.querySelector('code')?.textContent
  if (!code) return
  try {
    await navigator.clipboard.writeText(code)
    button.textContent = 'Copied!'
    window.setTimeout(() => { button.textContent = 'Copy' }, 1800)
  } catch {
    button.textContent = 'Select text to copy'
    window.setTimeout(() => { button.textContent = 'Copy' }, 2200)
  }
}))

document.getElementById('year').textContent = String(new Date().getFullYear())
