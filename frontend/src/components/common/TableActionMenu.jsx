import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

const TableActionMenu = ({ items = [], label = 'Open actions' }) => {
  const [open, setOpen] = useState(false)
  const menuRef = useRef(null)

  useEffect(() => {
    if (!open) return undefined

    const closeMenu = (event) => {
      if (menuRef.current?.contains(event.target)) return
      setOpen(false)
    }

    document.addEventListener('mousedown', closeMenu)
    return () => document.removeEventListener('mousedown', closeMenu)
  }, [open])

  const close = () => setOpen(false)

  return (
    <span className="table-action-cell" ref={menuRef}>
      <button
        className="action-menu-button"
        type="button"
        aria-label={label}
        aria-expanded={open}
        onClick={() => setOpen((current) => !current)}
      >
        <i className="bi bi-list"></i>
      </button>
      {open && (
        <div className="table-action-menu">
          {items.map((item) => {
            if (!item || item.hidden) return null
            const className = `table-action-item${item.danger ? ' danger' : ''}`
            if (item.to) {
              return (
                <Link key={item.label} className={className} to={item.to} onClick={close}>
                  {item.label}
                </Link>
              )
            }
            return (
              <button
                key={item.label}
                className={className}
                type="button"
                disabled={item.disabled}
                onClick={() => {
                  close()
                  item.onClick?.()
                }}
              >
                {item.label}
              </button>
            )
          })}
        </div>
      )}
    </span>
  )
}

export default TableActionMenu
