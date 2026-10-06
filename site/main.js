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

const tabs = [...document.querySelectorAll('[role="tab"]')]
tabs.forEach((tab, index) => {
  tab.addEventListener('click', () => selectTab(index))
  tab.addEventListener('keydown', event => {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
    event.preventDefault()
    const next = (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length
    selectTab(next)
    tabs[next].focus()
  })
})

function selectTab(index) {
  tabs.forEach((tab, current) => {
    const selected = current === index
    tab.setAttribute('aria-selected', String(selected))
    tab.tabIndex = selected ? 0 : -1
    document.getElementById(tab.getAttribute('aria-controls')).hidden = !selected
  })
}

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
