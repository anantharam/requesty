import { useState, useRef, useEffect } from 'react'

interface Model {
  id: string
  name: string
}

interface ModelSelectorProps {
  models: Model[]
  selectedModel: string
  onModelSelect: (modelId: string) => void
  disabled?: boolean
}

export default function ModelSelector({ 
  models, 
  selectedModel, 
  onModelSelect, 
  disabled = false 
}: ModelSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const selectedModelData = models.find(model => model.id === selectedModel)

  return (
    <div className="model-selector" ref={dropdownRef}>
      <div className="model-selector-header">
        <span className="model-selector-icon">🤖</span>
        <h2>Model Selector</h2>
      </div>
      <div className="custom-dropdown">
        <button
          className={`dropdown-button ${isOpen ? 'active' : ''} ${disabled ? 'disabled' : ''}`}
          onClick={() => !disabled && setIsOpen(!isOpen)}
          aria-expanded={isOpen}
          disabled={disabled}
        >
          <span className="selected-model">
            {selectedModelData?.name || 'Select Model'}
          </span>
          <span className={`arrow-icon ${isOpen ? 'up' : 'down'}`}>▼</span>
        </button>
        {isOpen && !disabled && (
          <div className="dropdown-menu" role="listbox">
            {models.map(model => (
              <button
                key={model.id}
                className={`dropdown-item ${model.id === selectedModel ? 'selected' : ''}`}
                onClick={() => {
                  onModelSelect(model.id)
                  setIsOpen(false)
                }}
                role="option"
                aria-selected={model.id === selectedModel}
              >
                <span className="model-name">{model.name}</span>
                {model.id === selectedModel && (
                  <span className="check-icon">✓</span>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
} 