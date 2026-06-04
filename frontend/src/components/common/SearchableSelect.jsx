import { useEffect, useMemo, useRef, useState } from 'react'

const SearchableSelect = ({
  value,
  options,
  onChange,
  placeholder = 'Select option',
  searchPlaceholder = 'Search...',
  emptyLabel = 'No options found',
}) => {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const wrapperRef = useRef(null)

  const selected = options.find((option) => option.value === value)
  const filteredOptions = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    if (!keyword) return options
    return options.filter((option) => (
      `${option.label} ${option.description || ''} ${(option.meta || []).map((item) => item.value).join(' ')}`.toLowerCase().includes(keyword)
    ))
  }, [options, query])

  useEffect(() => {
    const close = (event) => {
      if (!wrapperRef.current?.contains(event.target)) setOpen(false)
    }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [])

  const choose = (option) => {
    onChange(option.value)
    setQuery('')
    setOpen(false)
  }

  return (
    <div className="searchable-select" ref={wrapperRef}>
      <button
        type="button"
        className={`searchable-select-trigger ${open ? 'active' : ''}`}
        onClick={() => setOpen((next) => !next)}
      >
        <span>
          {selected ? selected.label : placeholder}
          {selected?.description && <small>{selected.description}</small>}
          {selected?.meta?.length > 0 && (
            <span className="searchable-select-meta">
              {selected.meta.slice(0, 2).map((item) => (
                <em key={`${item.label}-${item.value}`}>{item.label}: {item.value}</em>
              ))}
            </span>
          )}
        </span>
        <span className="searchable-select-arrow">
          <i className={`bi ${open ? 'bi-chevron-up' : 'bi-chevron-down'}`}></i>
        </span>
      </button>

      {open && (
        <div className="searchable-select-menu">
          <div className="searchable-select-input">
            <i className="bi bi-search"></i>
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={searchPlaceholder}
              autoFocus
            />
          </div>
          <div className="searchable-select-options">
            {filteredOptions.map((option) => (
              <button
                type="button"
                className={`searchable-select-option ${option.value === value ? 'selected' : ''}`}
                key={option.value}
                onClick={() => choose(option)}
              >
                <span>{option.label}</span>
                {option.description && <small>{option.description}</small>}
                {option.meta?.length > 0 && (
                  <span className="searchable-select-meta">
                    {option.meta.map((item) => (
                      <em key={`${item.label}-${item.value}`}>{item.label}: {item.value}</em>
                    ))}
                  </span>
                )}
              </button>
            ))}
            {!filteredOptions.length && <div className="searchable-select-empty">{emptyLabel}</div>}
          </div>
        </div>
      )}
    </div>
  )
}

export default SearchableSelect
