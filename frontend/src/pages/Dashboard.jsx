import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import dashboardService from '../services/dashboardService'
import walletService from '../services/walletService'
import providerService from '../services/providerService'
import { getErrorMessage } from '../services/api'
import { useAuth } from '../contexts/AuthContext'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const Dashboard = () => {
  const { hasPermission } = useAuth()
  const [summary, setSummary] = useState(null)
  const [wallet, setWallet] = useState(null)
  const [providerBalance, setProviderBalance] = useState(null)
  const [loading, setLoading] = useState(true)
  const canViewProviderData = hasPermission('provider_balance', 'view')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const requests = [
        dashboardService.getSummary('smm'),
        walletService.getMyWallet(),
      ]
      if (canViewProviderData) requests.push(providerService.getBalance())

      const results = await Promise.allSettled(requests)
      if (results[0].status === 'fulfilled') setSummary(results[0].value.data.data)
      if (results[1].status === 'fulfilled') setWallet(results[1].value.data.data)
      if (canViewProviderData && results[2]?.status === 'fulfilled') {
        setProviderBalance(results[2].value.data.data)
      }
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load dashboard'))
    } finally {
      setLoading(false)
    }
  }, [canViewProviderData])

  useEffect(() => {
    load()
  }, [load])

  const cards = [
    { label: 'Total Orders', value: summary?.total_orders, icon: 'bi-bag-heart', color: 'primary' },
    { label: 'Completed', value: summary?.completed_orders, icon: 'bi-check2-circle', color: 'success' },
    { label: 'Processing', value: summary?.processing_orders, icon: 'bi-arrow-repeat', color: 'warning' },
    { label: 'Pending', value: summary?.pending_orders, icon: 'bi-hourglass-split', color: 'info' },
  ]

  return (
    <div className="luxe-page-fade">
      <div className="luxe-detail-hero mb-4">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
             <i className="bi bi-speedometer2"></i> Operation Center
          </div>
          <h1 className="luxe-hero-title">Overview</h1>
          <p className="luxe-hero-subtitle">
            {canViewProviderData 
               ? 'Monitor SMM operations, profit margins, and provider readiness.' 
               : 'Track your personal SMM orders and wallet balance activity.'}
          </p>
        </div>
        <div className="toolbar-actions">
           <button type="button" className="btn btn-outline-dark d-flex align-items-center gap-2" onClick={load}>
              <i className={`bi bi-arrow-repeat ${loading ? 'spin' : ''}`}></i> Refresh
           </button>
           <Link to="/smm/orders/new" className="btn btn-primary d-flex align-items-center gap-2">
              <i className="bi bi-plus-circle-fill"></i> New Order
           </Link>
        </div>
      </div>

      <div className="metric-grid mb-4">
        {cards.map((card) => (
          <div className="metric-card" key={card.label}>
            <div className={`metric-icon-box ${card.color}`}>
               <i className={`bi ${card.icon}`}></i>
            </div>
            <div className="metric-info">
               <span>{card.label}</span>
               <strong>{loading ? '...' : new Intl.NumberFormat('id-ID').format(card.value || 0)}</strong>
            </div>
          </div>
        ))}
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-graph-up-arrow"></i> Financial Snapshot</h3>
            <span className="status-badge status-badge-sm info">Live</span>
          </div>
          <div className="luxe-card-body">
            <div className="money-row">
              <span>Total Spending</span>
              <strong>{formatMoney(summary?.total_amount)}</strong>
            </div>
            {canViewProviderData && (
              <>
                <div className="money-row">
                  <span>Provider Cost</span>
                  <span className="text-muted">{formatMoney(summary?.total_provider_charge)}</span>
                </div>
                <div className="money-row highlight mt-3 pt-3 border-top">
                  <span>Net Profit</span>
                  <strong className="text-success" style={{ fontSize: '20px' }}>{formatMoney(summary?.total_profit)}</strong>
                </div>
              </>
            )}
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-wallet2"></i> Current Balances</h3>
            <span className={`status-badge status-badge-sm ${wallet?.is_active !== false ? 'success' : 'danger'}`}>
               {wallet?.is_active !== false ? 'Active' : 'Locked'}
            </span>
          </div>
          <div className="luxe-card-body">
            <div className="money-row">
              <span>Your Wallet</span>
              <strong className="text-primary" style={{ fontSize: '20px' }}>{formatMoney(wallet?.balance)}</strong>
            </div>
            <div className="money-row">
              <span>Currency</span>
              <strong>{wallet?.currency || 'IDR'}</strong>
            </div>
            {canViewProviderData && (
              <div className="money-row mt-3 pt-3 border-top">
                <span>H2H Provider</span>
                <strong>{formatMoney(providerBalance?.balance || providerBalance?.data?.balance)}</strong>
              </div>
            )}
          </div>
        </section>
      </div>

      <h3 className="mt-4 mb-3" style={{ fontSize: '18px', fontWeight: 700, color: 'var(--ink)' }}>Quick Navigation</h3>
      <div className="quick-actions">
        <Link to="/smm/services" className="luxe-card-header py-3 px-4" style={{ borderRadius: '16px', border: '1px solid var(--hairline-soft)' }}>
           <i className="bi bi-grid-fill text-primary"></i> Browse Services
        </Link>
        <Link to="/smm/orders" className="luxe-card-header py-3 px-4" style={{ borderRadius: '16px', border: '1px solid var(--hairline-soft)' }}>
           <i className="bi bi-bag-check-fill text-success"></i> Review Orders
        </Link>
        <Link to="/deposits" className="luxe-card-header py-3 px-4" style={{ borderRadius: '16px', border: '1px solid var(--hairline-soft)' }}>
           <i className="bi bi-receipt-cutoff text-warning"></i> Add Funds
        </Link>
        <Link to="/wallet" className="luxe-card-header py-3 px-4" style={{ borderRadius: '16px', border: '1px solid var(--hairline-soft)' }}>
           <i className="bi bi-journals text-info"></i> Transaction Log
        </Link>
      </div>
    </div>
  )
}

export default Dashboard
