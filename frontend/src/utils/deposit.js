export const depositMetadata = (deposit) => {
  if (!deposit?.metadata) return {}
  if (typeof deposit.metadata === 'string') {
    try {
      return JSON.parse(deposit.metadata)
    } catch {
      return {}
    }
  }
  return deposit.metadata
}

export const isQRISDeposit = (deposit) => {
  const metadata = depositMetadata(deposit)
  return ['qris', 'qrisly'].includes(deposit?.provider) || ['qris', 'qrisly'].includes(metadata.payment_channel)
}

export const depositPayableAmount = (deposit) => depositMetadata(deposit).payable_amount || deposit?.amount

export const qrisImageURL = (deposit) => {
  const metadata = depositMetadata(deposit)
  const value = deposit?.payment_url || metadata.qris_image_url || metadata.qris_image || metadata.qris_image_base64 || ''
  if (!value) return ''
  if (value.startsWith('http') || value.startsWith('data:image')) return value
  if (/^[A-Za-z0-9+/=\s]+$/.test(value) && value.length > 120) {
    return `data:image/png;base64,${value.replace(/\s/g, '')}`
  }
  return value
}

export const qrisString = (deposit) => depositMetadata(deposit).qris_string || ''

export const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))
