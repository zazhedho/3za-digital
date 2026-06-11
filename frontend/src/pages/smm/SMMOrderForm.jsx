import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import SearchableSelect from '../../components/common/SearchableSelect'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const getMinQuantity = (service) => Math.max(1, Number(service?.min_quantity || 1))

const getMaxQuantity = (service) => {
  const max = Number(service?.max_quantity || 0)
  return max > 0 ? max : undefined
}

const clampQuantity = (value, service) => {
  const min = getMinQuantity(service)
  const max = getMaxQuantity(service)
  const quantity = Math.trunc(Number(value))

  if (!Number.isFinite(quantity) || quantity < min) return min
  if (max && quantity > max) return max
  return quantity
}

const platformOptions = [
  { value: '', label: 'All platforms' },
  { value: 'instagram', label: 'Instagram' },
  { value: 'tiktok', label: 'TikTok' },
  { value: 'youtube', label: 'YouTube' },
  { value: 'facebook', label: 'Facebook' },
]

const SMMOrderForm = () => {
  const [services, setServices] = useState([])
  const [serviceSearch, setServiceSearch] = useState('')
  const [serviceLoading, setServiceLoading] = useState(false)
  const [platform, setPlatform] = useState('')
  const [form, setForm] = useState({ service_id: '', target: '', quantity: 1 })
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const selectedService = services.find((service) => service.id === form.service_id)
  const minQuantity = getMinQuantity(selectedService)
  const maxQuantity = getMaxQuantity(selectedService)

  const loadServices = useCallback(async (keyword = '') => {
    setServiceLoading(true)
    try {
      const response = await smmService.getServices({
        search: keyword,
        limit: 50,
        'filters[platform]': platform || undefined,
        'filters[is_active]': 'true',
      })
      setServices(getListPayload(response).rows)
    } catch {
      setServices([])
    } finally {
      setServiceLoading(false)
    }
  }, [platform])

  useEffect(() => {
    const timer = setTimeout(() => {
      loadServices(serviceSearch)
    }, 300)
    return () => clearTimeout(timer)
  }, [loadServices, serviceSearch])

  const handleServiceChange = (serviceId) => {
    const service = services.find((item) => item.id === serviceId)
    setForm((current) => ({
      ...current,
      service_id: serviceId,
      quantity: getMinQuantity(service),
    }))
  }

  const handlePlatformChange = (event) => {
    setPlatform(event.target.value)
    setServiceSearch('')
    setForm((current) => ({ ...current, service_id: '', quantity: 1 }))
  }

  const handleQuantityChange = (event) => {
    const value = event.target.value
    // Only allow numeric input but don't clamp yet
    if (value === '' || /^\d+$/.test(value)) {
      setForm((current) => ({ ...current, quantity: value }))
    }
  }

  const handleQuantityBlur = () => {
    setForm((current) => ({
      ...current,
      quantity: clampQuantity(current.quantity, selectedService),
    }))
  }

  const submit = async (event) => {
    event.preventDefault()
    if (!form.service_id) {
      toast.error('Service is required')
      return
    }
    const quantity = clampQuantity(form.quantity, selectedService)
    setLoading(true)
    try {
      await smmService.createOrder({ ...form, quantity })
      toast.success('Order created')
      navigate('/smm/orders')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to create order'))
    } finally {
      setLoading(false)
    }
  }

  const estimatedCost = selectedService ? (Number(form.quantity || 0) / 1000) * selectedService.price : 0

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Create SMM Order</h1>
          <p>Boost your social media presence with our premium services.</p>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-plus-circle"></i> New Order</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={submit} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Platform</label>
                <div className="deposit-select-wrapper">
                  <i className="bi bi-share"></i>
                  <select value={platform} onChange={handlePlatformChange}>
                    {platformOptions.map((option) => (
                      <option key={option.value || 'all'} value={option.value}>{option.label}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="deposit-input-group">
                <label>Service</label>
                <SearchableSelect
                  key={platform || 'all-platforms'}
                  value={form.service_id}
                  onChange={handleServiceChange}
                  onSearch={setServiceSearch}
                  loading={serviceLoading}
                  placeholder="Select service"
                  searchPlaceholder="Search service name or provider ID"
                  emptyLabel={serviceSearch ? 'No matching service found' : 'No active services found'}
                  options={services.map((service) => ({
                    value: service.id,
                    label: service.name,
                    description: service.category || service.brand || 'SMM service',
                    meta: [
                      { label: 'ID', value: service.provider_service_id || '-' },
                      { label: 'Platform', value: service.platform || '-' },
                      { label: 'Price/1k', value: formatMoney(service.price) },
                      { label: 'Qty', value: `${getMinQuantity(service)}-${getMaxQuantity(service) || '-'}` },
                    ],
                  }))}
                />
              </div>

              <div className="deposit-input-group">
                <label>Target URL / Link</label>
                <div className="auth-input m-0">
                   <i className="bi bi-link-45deg"></i>
                   <input
                    type="url"
                    pattern="https?://.+"
                    placeholder="https://instagram.com/username"
                    value={form.target}
                    onChange={(event) => setForm({ ...form, target: event.target.value })}
                    required
                    style={{ background: 'white' }}
                  />
                </div>
              </div>

              <div className="deposit-input-group">
                <label>Quantity</label>
                <div className="deposit-amount-wrapper">
                  <span className="deposit-amount-currency"><i className="bi bi-hash"></i></span>
                  <input
                    type="number"
                    min={minQuantity}
                    max={maxQuantity}
                    value={form.quantity}
                    onChange={handleQuantityChange}
                    onBlur={handleQuantityBlur}
                    required
                    style={{ fontSize: '18px', height: '52px', paddingLeft: '44px' }}
                  />
                </div>
                {selectedService && (
                  <div className="form-hint">
                    Min {minQuantity.toLocaleString('id-ID')}{maxQuantity ? ` / Max ${maxQuantity.toLocaleString('id-ID')}` : ''}
                  </div>
                )}
              </div>

              <div className="toolbar-actions justify-content-end mt-4 pt-3 border-top">
                <button className="btn btn-outline-dark" type="button" onClick={() => navigate('/smm/orders')}>Cancel</button>
                <button className="btn btn-primary px-4" disabled={loading}>
                  {loading ? 'Creating...' : 'Submit Order'}
                </button>
              </div>
            </form>
          </div>
        </section>

        <section className="luxe-detail-card h-fit">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Summary & Guide</h3>
          </div>
          <div className="luxe-card-body">
             <div className="deposit-fee-card mb-4">
                <div className="deposit-fee-row">
                   <span>Service Price</span>
                   <strong>{selectedService ? `${formatMoney(selectedService.price)}/1k` : '-'}</strong>
                </div>
                <div className="deposit-fee-row total">
                   <span>Estimated Cost</span>
                   <strong>{selectedService ? formatMoney(estimatedCost) : '-'}</strong>
                </div>
             </div>
             
             <div className="luxe-item mb-3">
                <span className="luxe-label">Delivery Speed</span>
                <span className="luxe-value">Usually starts within 0-24 hours depending on the service quality.</span>
             </div>
             <div className="luxe-item">
                <span className="luxe-label">Important Note</span>
                <span className="luxe-value">Please make sure your target profile is <strong>Public</strong>. Orders to private accounts will fail without refund.</span>
             </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default SMMOrderForm
