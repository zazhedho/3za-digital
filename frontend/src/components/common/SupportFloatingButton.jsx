import { useEffect, useState, useRef } from 'react'
import supportService from '../../services/supportService'
import '../../styles/SupportFloatingButton.css'

const SupportFloatingButton = () => {
  const [contact, setContact] = useState(null)
  const [isOpen, setIsOpen] = useState(false)
  const wrapperRef = useRef(null)

  useEffect(() => {
    supportService.getSupportContact()
      .then((response) => {
        setContact(response.data.data)
      })
      .catch(() => {
        setContact(null)
      })
  }, [])

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  if (!contact || (!contact.whatsapp && !contact.telegram && !contact.email)) return null

  const whatsappUrl = contact.whatsapp ? `https://wa.me/${contact.whatsapp.replace(/\D/g, '')}` : null

  return (
    <div className={`support-hub-wrapper ${isOpen ? 'is-open' : ''}`} ref={wrapperRef}>
      {/* Support Menu Panel */}
      <div className="support-panel">
        <div className="support-panel-header">
          <strong>Support Center</strong>
          <p>Choose your preferred channel</p>
        </div>
        
        <div className="support-options">
          {whatsappUrl && (
            <a href={whatsappUrl} target="_blank" rel="noopener noreferrer" className="support-opt-item wa">
              <div className="opt-icon"><i className="bi bi-whatsapp"></i></div>
              <div className="opt-text">
                <span>WhatsApp</span>
                <small>Fast Response</small>
              </div>
            </a>
          )}
          
          {contact.telegram && (
            <a href={`https://t.me/${contact.telegram}`} target="_blank" rel="noopener noreferrer" className="support-opt-item tg">
              <div className="opt-icon"><i className="bi bi-telegram"></i></div>
              <div className="opt-text">
                <span>Telegram</span>
                <small>Community & Support</small>
              </div>
            </a>
          )}
          
          {contact.email && (
            <a href={`mailto:${contact.email}`} className="support-opt-item mail">
              <div className="opt-icon"><i className="bi bi-envelope-at"></i></div>
              <div className="opt-text">
                <span>Email</span>
                <small>Business Inquiries</small>
              </div>
            </a>
          )}
        </div>
      </div>

      {/* Main Toggle Button (Icon Only) */}
      <button
        type="button"
        className="support-hub-trigger-circle"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Toggle support menu"
      >
        <div className="trigger-icon-stack">
          <i className={`bi ${isOpen ? 'bi-x-lg' : 'bi-headset'} icon-main`}></i>
        </div>
      </button>
    </div>
  )
}

export default SupportFloatingButton
