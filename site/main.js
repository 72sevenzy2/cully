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

const testimonialTrack = document.getElementById('testimonial-track')
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
const testimonialButtons = document.querySelectorAll('[data-testimonial-direction]')

function updateTestimonialControls() {
  if (!testimonialTrack) return
  const remaining = testimonialTrack.scrollWidth - testimonialTrack.clientWidth - testimonialTrack.scrollLeft
  testimonialButtons.forEach(button => {
    button.disabled = Number(button.dataset.testimonialDirection) < 0
      ? testimonialTrack.scrollLeft <= 1 : remaining <= 1
  })
}

testimonialButtons.forEach(button => button.addEventListener('click', () => {
  const card = testimonialTrack.querySelector('.testimonial-card')
  const distance = card ? card.getBoundingClientRect().width + 24 : testimonialTrack.clientWidth
  testimonialTrack.scrollBy({ left: Number(button.dataset.testimonialDirection) * distance,
    behavior: reducedMotion.matches ? 'instant' : 'smooth' })
}))
testimonialTrack?.addEventListener('scroll', updateTestimonialControls, { passive: true })
window.addEventListener('resize', updateTestimonialControls)
updateTestimonialControls()

async function loadTestimonials() {
  if (!testimonialTrack) return
  try {
    const response = await fetch('/testimonials.json')
    if (!response.ok) return
    const entries = await response.json()
    if (!Array.isArray(entries)) return
    const approved = entries.filter(entry => entry && typeof entry.name === 'string'
      && entry.name.trim() && typeof entry.quote === 'string' && entry.quote.trim())
    if (!approved.length) return
    const cards = approved.map(entry => {
      const card = document.createElement('article')
      card.className = 'testimonial-card'
      const mark = document.createElement('span')
      mark.className = 'quote-mark'
      mark.setAttribute('aria-hidden', 'true')
      mark.textContent = '“'
      const quote = document.createElement('blockquote')
      quote.textContent = entry.quote
      const person = document.createElement('div')
      person.className = 'testimonial-person'
      const avatar = document.createElement('span')
      avatar.className = 'testimonial-avatar'
      avatar.setAttribute('aria-hidden', 'true')
      avatar.textContent = entry.name.trim().slice(0, 1).toUpperCase()
      const photoURL = typeof entry.photoUrl === 'string' && /^\/testimonial-photos\/[a-f0-9]{32}$/.test(entry.photoUrl)
        ? new URL(entry.photoUrl, window.location.origin) : safeHTTPSURL(entry.photoUrl)
      if (photoURL) {
        const photo = document.createElement('img')
        photo.src = photoURL.href
        photo.alt = ''
        photo.loading = 'lazy'
        photo.addEventListener('error', () => { avatar.textContent = entry.name.trim().slice(0, 1).toUpperCase() })
        avatar.replaceChildren(photo)
      }
      const details = document.createElement('div')
      const name = document.createElement('strong')
      const linkedinURL = safeLinkedInURL(entry.linkedin)
      if (linkedinURL) {
        const profile = document.createElement('a')
        profile.href = linkedinURL.href
        profile.target = '_blank'
        profile.rel = 'noopener noreferrer'
        profile.textContent = entry.name
        name.append(profile)
      } else {
        name.textContent = entry.name
      }
      const context = document.createElement('small')
      context.textContent = [entry.context, entry.workplace].filter(value => typeof value === 'string' && value.trim()).join(' · ') || 'Cully community'
      details.append(name, context)
      person.append(avatar, details)
      card.append(mark, quote, person)
      return card
    })
    testimonialTrack.replaceChildren(...cards)
    document.getElementById('testimonial-caption').textContent = 'Experiences shared by the Cully community.'
    updateTestimonialControls()
  } catch {
    // Keep the invitation cards available if the content cannot be loaded.
  }
}
loadTestimonials()

function safeHTTPSURL(value) {
  try {
    const url = new URL(value)
    return url.protocol === 'https:' && !url.username && !url.password ? url : null
  } catch {
    return null
  }
}

function safeLinkedInURL(value) {
  const url = safeHTTPSURL(value)
  return url && ['linkedin.com', 'www.linkedin.com'].includes(url.hostname)
    && url.pathname.startsWith('/in/') && url.pathname.length > 4 ? url : null
}

const photoInput = document.getElementById('testimonial-photo')
let photoPreviewURL
photoInput?.addEventListener('change', () => {
  if (photoPreviewURL) URL.revokeObjectURL(photoPreviewURL)
  const photo = photoInput.files[0]
  const preview = document.getElementById('testimonial-photo-preview')
  const placeholder = document.getElementById('testimonial-photo-placeholder')
  const status = document.getElementById('testimonial-photo-status')
  preview.hidden = true
  preview.removeAttribute('src')
  placeholder.hidden = false
  photoInput.setCustomValidity('')
  document.querySelector('input[name=importedPhoto]').value = ''
  status.textContent = 'Add a photo to put a face to your story.'
  if (!photo) return
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(photo.type) || photo.size > 5 * 1024 * 1024) {
    photoInput.setCustomValidity('Choose a JPG, PNG, or WebP image up to 5 MB.')
    status.textContent = photoInput.validationMessage
    photoInput.reportValidity()
    return
  }
  photoPreviewURL = URL.createObjectURL(photo)
  preview.src = photoPreviewURL
  preview.hidden = false
  placeholder.hidden = true
  status.textContent = `${photo.name} selected. Your photo will be uploaded with your testimonial.`
})

const linkedinInput = document.querySelector('#share-testimonial input[name=linkedin]')
linkedinInput?.addEventListener('input', () => {
  linkedinInput.setCustomValidity(linkedinInput.value.trim() && !safeLinkedInURL(linkedinInput.value.trim())
    ? 'Enter a LinkedIn profile URL such as https://www.linkedin.com/in/your-name.' : '')
})

document.getElementById('import-linkedin')?.addEventListener('click', async event => {
  const button = event.currentTarget
  const status = document.getElementById('linkedin-import-status')
  const url = linkedinInput.value.trim()
  if (!safeLinkedInURL(url)) {
    linkedinInput.setCustomValidity('Enter a LinkedIn profile URL first.')
    linkedinInput.reportValidity()
    return
  }
  button.disabled = true
  status.textContent = 'Checking public profile details…'
  try {
    const response = await fetch('/api/linkedin-profile', {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-Cully-Submission': '1' },
      body: JSON.stringify({ url }), signal: AbortSignal.timeout(15000)
    })
    const profile = await response.json()
    if (!response.ok) throw new Error(profile.error || 'Could not import this profile. Fill in your details manually.')
    const form = document.getElementById('share-testimonial')
    for (const name of ['name', 'context', 'workplace']) {
      const input = form.elements.namedItem(name)
      if (!input.value.trim() && typeof profile[name] === 'string') input.value = profile[name].slice(0, input.maxLength)
    }
    const photoURL = safeHTTPSURL(profile.photoUrl)
    if (photoURL?.hostname === 'media.licdn.com' && !photoInput.files.length) {
      form.elements.namedItem('importedPhoto').value = photoURL.href
      const preview = document.getElementById('testimonial-photo-preview')
      preview.src = photoURL.href
      preview.hidden = false
      document.getElementById('testimonial-photo-placeholder').hidden = true
      document.getElementById('testimonial-photo-status').textContent = 'Public LinkedIn photo selected. You can upload a different photo instead.'
    }
    status.textContent = 'Public details imported. Please check and edit them before submitting.'
  } catch (error) {
    status.textContent = error.name === 'TimeoutError' ? 'LinkedIn took too long. Fill in your details manually.' : error.message
  } finally {
    button.disabled = false
  }
})

document.getElementById('share-testimonial')?.addEventListener('submit', async event => {
  event.preventDefault()
  const form = event.currentTarget
  const data = new FormData(form)
  const name = String(data.get('name') || '').trim()
  const quote = String(data.get('quote') || '').trim()
  if (!name || quote.length < 20) {
    form.elements.namedItem(!name ? 'name' : 'quote').focus()
    return
  }
  const button = form.querySelector('button[type=submit]')
  const status = document.getElementById('testimonial-submit-status')
  button.disabled = true
  status.textContent = 'Uploading your testimonial…'
  try {
    const response = await fetch('/api/testimonials', { method: 'POST',
      headers: { 'X-Cully-Submission': '1' }, body: data, signal: AbortSignal.timeout(40000) })
    const result = await response.json()
    if (!response.ok) throw new Error(result.error || 'Could not save your testimonial. Please retry.')
    form.reset()
    photoInput.dispatchEvent(new Event('change'))
    linkedinInput.setCustomValidity('')
    document.getElementById('linkedin-import-status').textContent = ''
    status.textContent = result.message
  } catch (error) {
    status.textContent = error.name === 'TimeoutError' ? 'The upload timed out. Please retry.' : error.message
  } finally {
    button.disabled = false
  }
})
