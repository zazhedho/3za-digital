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

export const depositProviderLabel = (deposit) => {
  const metadata = depositMetadata(deposit)
  const provider = deposit?.provider || metadata.payment_channel
  const qrisType = metadata.qris_type
  if (provider === 'qrisly' || qrisType === 'dynamic') return 'Dynamic QRIS'
  if (provider === 'qris' || qrisType === 'static') return 'Static QRIS'
  if (provider === 'manual') return 'Manual Review'
  return provider || '-'
}

export const depositStatus = (deposit) => {
  const metadata = depositMetadata(deposit)
  if (isQRISDeposit(deposit) && deposit?.status === 'pending' && metadata.qrisly_status) {
    return metadata.qrisly_status
  }
  return deposit?.status || '-'
}

export const depositStatusClass = (deposit) => {
  const status = String(depositStatus(deposit)).toLowerCase()
  if (['paid', 'success', 'settlement'].includes(status)) return 'status-paid'
  if (['unpaid'].includes(status)) return 'status-unpaid'
  if (['pending'].includes(status)) return 'status-pending'
  if (['expired'].includes(status)) return 'status-expired'
  if (['failed', 'deny'].includes(status)) return 'status-failed'
  if (['cancelled', 'canceled', 'cancel'].includes(status)) return 'status-cancelled'
  return 'status-default'
}

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
