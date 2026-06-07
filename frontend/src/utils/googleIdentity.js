const googleScriptId = 'google-identity-services'

export const getGoogleClientId = () => String(
  window.ENV_CONFIG?.GOOGLE_CLIENT_ID || import.meta.env.VITE_GOOGLE_CLIENT_ID || '',
).trim()

export const renderGoogleIdentityButton = ({ element, clientId, onCredential, text = 'signin_with' }) => {
  if (!element || !clientId) return undefined

  const renderButton = () => {
    if (!window.google?.accounts?.id || !element) return
    element.innerHTML = ''
    window.google.accounts.id.initialize({
      client_id: clientId,
      callback: (response) => onCredential(response.credential),
    })
    window.google.accounts.id.renderButton(element, {
      theme: 'outline',
      size: 'large',
      width: element.offsetWidth || 320,
      text,
    })
  }

  if (window.google?.accounts?.id) {
    renderButton()
    return undefined
  }

  let script = document.getElementById(googleScriptId)
  if (!script) {
    script = document.createElement('script')
    script.id = googleScriptId
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    document.body.appendChild(script)
  }

  script.addEventListener('load', renderButton)
  return () => script.removeEventListener('load', renderButton)
}
