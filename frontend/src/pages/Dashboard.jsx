import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import dashboardService from '../services/dashboardService'
import walletService from '../services/walletService'
import providerService from '../services/providerService'
import { getErrorMessage } from '../services/api'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const Dashboard = () => {
  const [summary, setSummary] = useState(null)
  const [wallet, setWallet] = useState(null)
  const [providerBalance, setProviderBalance] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const [summaryRes, walletRes, providerRes] = await Promise.allSettled([
        dashboardService.getSummary('smm'),
        walletService.getMyWallet(),
        providerService.getBalance(),
      ])
      if (summaryRes.status === 'fulfilled') setSummary(summaryRes.value.data.data)
      if (walletRes.status === 'fulfilled') setWallet(walletRes.value.data.data)
      if (providerRes.status === 'fulfilled') setProviderBalance(providerRes.value.data.data)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load dashboard'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const cards = [
    { label: 'Total orders', value: summary?.total_orders, icon: 'bi-bag' },
    { label: 'Completed', value: summary?.completed_orders, icon: 'bi-check2-circle' },
    { label: 'Processing', value: summary?.processing_orders, icon: 'bi-arrow-repeat' },
    { label: 'Pending', value: summary?.pending_orders, icon: 'bi-hourglass-split' },
  ]

  return (
    <div>
      <div className="page-hero">
        <div>
          <h1>3ZA Digital Operations</h1>
          <p>Track SMM orders, margin, wallet balance, and provider readiness.</p>
        </div>
        <div className="hero-actions">
          <Link to="/smm/orders/new" className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>New order</Link>
          <button type="button" className="btn btn-outline-dark" onClick={load}>Refresh</button>
        </div>
      </div>

      <div className="metric-grid">
        {cards.map((card) => (
          <div className="metric-card" key={card.label}>
            <i className={`bi ${card.icon}`}></i>
            <span>{card.label}</span>
            <strong>{loading ? '-' : new Intl.NumberFormat('id-ID').format(card.value || 0)}</strong>
          </div>
        ))}
      </div>

      <div className="content-grid two">
        <section className="panel">
          <div className="panel-heading">
            <h2>Revenue snapshot</h2>
            <span className="status-dot">SMM</span>
          </div>
          <div className="money-row">
            <span>Total amount</span>
            <strong>{formatMoney(summary?.total_amount)}</strong>
          </div>
          <div className="money-row">
            <span>Provider charge</span>
            <strong>{formatMoney(summary?.total_provider_charge)}</strong>
          </div>
          <div className="money-row highlight">
            <span>Profit</span>
            <strong>{formatMoney(summary?.total_profit)}</strong>
          </div>
        </section>

        <section className="panel">
          <div className="panel-heading">
            <h2>Balances</h2>
            <span className="status-dot">Live</span>
          </div>
          <div className="money-row">
            <span>My wallet</span>
            <strong>{formatMoney(wallet?.balance)}</strong>
          </div>
          <div className="money-row">
            <span>Currency</span>
            <strong>{wallet?.currency || 'IDR'}</strong>
          </div>
          <div className="money-row">
            <span>Provider</span>
            <strong>{formatMoney(providerBalance?.balance || providerBalance?.data?.balance)}</strong>
          </div>
        </section>
      </div>

      <div className="quick-actions">
        <Link to="/smm/services"><i className="bi bi-grid"></i>Browse services</Link>
        <Link to="/smm/orders"><i className="bi bi-bag-check"></i>Review orders</Link>
        <Link to="/deposits"><i className="bi bi-receipt"></i>Deposit history</Link>
        <Link to="/wallet"><i className="bi bi-wallet2"></i>Wallet ledger</Link>
      </div>
    </div>
  )
}

export default Dashboard
