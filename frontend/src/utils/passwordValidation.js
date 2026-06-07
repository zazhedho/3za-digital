export const validatePassword = (value = '') => ({
  minLength: value.length >= 8,
  hasLowercase: /[a-z]/.test(value),
  hasUppercase: /[A-Z]/.test(value),
  hasNumber: /[0-9]/.test(value),
  hasSymbol: /[^a-zA-Z0-9]/.test(value),
})

export const passwordStrength = (validation) => Object.values(validation).filter(Boolean).length

export const isPasswordValid = (validation) => Object.values(validation).every(Boolean)

export const passwordStrengthLabel = (strength) => {
  if (!strength) return ''
  if (strength <= 2) return 'Weak'
  if (strength === 3) return 'Fair'
  if (strength === 4) return 'Good'
  return 'Strong'
}

export const passwordRequirements = [
  ['minLength', 'Min 8 chars'],
  ['hasLowercase', 'Lowercase'],
  ['hasUppercase', 'Uppercase'],
  ['hasNumber', 'Number'],
  ['hasSymbol', 'Symbol'],
]
