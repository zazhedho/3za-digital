import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Link } from 'react-router-dom'

const TableActionMenu = ({ items = [], label = 'Open actions' }) => {
  const [open, setOpen] = useState(false)
  const [openUp, setOpenUp] = useState(false)
  const [position, setPosition] = useState({ left: 0, top: 0 })
  const buttonRef = useRef(null)
  const menuRef = useRef(null)
  const visibleItems = items.filter((item) => item && !item.hidden)

  const updatePosition = useCallback(() => {
    const button = buttonRef.current
    if (!button) return

    const buttonRect = button.getBoundingClientRect()
    const menuWidth = menuRef.current?.offsetWidth || 168
    const menuHeight = menuRef.current?.offsetHeight || Math.max(44, visibleItems.length * 38)
    const gap = 6
    const viewportPadding = 10
    const spaceBelow = window.innerHeight - buttonRect.bottom
    const shouldOpenUp = spaceBelow < menuHeight + gap && buttonRect.top > menuHeight + gap
    const top = shouldOpenUp
      ? Math.max(viewportPadding, buttonRect.top - menuHeight - gap)
      : Math.min(window.innerHeight - menuHeight - viewportPadding, buttonRect.bottom + gap)
    const left = Math.min(
      window.innerWidth - menuWidth - viewportPadding,
      Math.max(viewportPadding, buttonRect.right - menuWidth),
    )

    setOpenUp(shouldOpenUp)
    setPosition({ left, top })
  }, [visibleItems.length])

  useLayoutEffect(() => {
    if (open) updatePosition()
  }, [open, updatePosition])

  useEffect(() => {
    if (!open) return undefined

    const closeMenu = (event) => {
      if (buttonRef.current?.contains(event.target)) return
      if (menuRef.current?.contains(event.target)) return
      setOpen(false)
    }

    document.addEventListener('mousedown', closeMenu)
    window.addEventListener('resize', updatePosition)
    window.addEventListener('scroll', updatePosition, true)
    return () => {
      document.removeEventListener('mousedown', closeMenu)
      window.removeEventListener('resize', updatePosition)
      window.removeEventListener('scroll', updatePosition, true)
    }
  }, [open, updatePosition])

  const close = () => setOpen(false)
  const menu = open ? createPortal(
    <div
      className={`table-action-menu table-action-menu-portal ${openUp ? 'open-up' : ''}`}
      ref={menuRef}
      style={{ left: position.left, top: position.top }}
    >
      {visibleItems.map((item) => {
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
    </div>,
    document.body,
  ) : null

  return (
    <span className="table-action-cell">
      <button
        className="action-menu-button"
        type="button"
        ref={buttonRef}
        aria-label={label}
        aria-expanded={open}
        onClick={() => setOpen((current) => !current)}
      >
        <i className="bi bi-list"></i>
      </button>
      {menu}
    </span>
  )
}

export default TableActionMenu
