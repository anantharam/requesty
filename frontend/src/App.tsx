import { useState, useRef, useEffect } from 'react'
import axios from 'axios'
import './App.css'
import ModelSelector from './components/ModelSelector'

interface Message {
  role: 'user' | 'assistant'
  content: string
  cost?: {
    total_cost: number
    model_info?: {
      id: string
    }
  }
}

const AVAILABLE_MODELS = [
  { id: 'gpt-4o-mini', name: 'GPT-4o Mini' },
  { id: 'gpt-4.1', name: 'GPT-4.1' },
  { id: 'gpt-4.5-preview', name: 'GPT-4.5 Preview' }
]

function App() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [selectedModel, setSelectedModel] = useState(AVAILABLE_MODELS[0].id)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  // Scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const sendMessage = async (stream: boolean) => {
    if (!input.trim()) return

    const userMessage: Message = {
      role: 'user',
      content: input
    }

    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsLoading(true)

    try {
      if (stream) {
        const response = await fetch('/api/v1/messages', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            messages: [...messages, userMessage],
            model: selectedModel,
            stream: true,
            max_tokens: 1000
          })
        })

        if (!response.body) throw new Error('No response body')

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        
        // Initialize assistant message
        const assistantMessage: Message = {
          role: 'assistant',
          content: ''
        }
        setMessages(prev => [...prev, assistantMessage])

        let lastResponse: any = null
        while (true) {
          const { value, done } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value)
          const lines = chunk.split('\n').filter(line => line.trim() !== '')
          
          for (const line of lines) {
            try {
              if (line === 'data: [DONE]') continue
              const streamResponse = JSON.parse(line)
              if (streamResponse.delta?.text) {
                assistantMessage.content += streamResponse.delta.text
                lastResponse = streamResponse
                setMessages(prev => 
                  prev.map((msg, i) => i === prev.length - 1 ? {...assistantMessage} : msg)
                )
              }
            } catch (e) {
              console.error('Error parsing chunk:', e)
            }
          }
        }

        // Update the last message with cost information if available
        if (lastResponse?.cost) {
          setMessages(prev => 
            prev.map((msg, i) => i === prev.length - 1 ? {...msg, cost: lastResponse.cost} : msg)
          )
        }
      } else {
        const response = await axios.post('/api/v1/messages', {
          messages: [...messages, userMessage],
          model: selectedModel,
          stream: false,
          max_tokens: 1000
        })

        const assistantMessage: Message = {
          role: 'assistant',
          content: response.data.content,
          cost: response.data.cost
        }

        setMessages(prev => [...prev, assistantMessage])
      }
    } catch (error) {
      console.error('Error sending message:', error)
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: 'Sorry, an error occurred while processing your message.'
      }])
    } finally {
      setIsLoading(false)
    }
  }

  const formatCost = (cost: Message['cost']) => {
    if (!cost?.model_info?.id) {
      return 'Model not found'
    }
    return `Cost: $${cost.total_cost.toFixed(6)}`
  }

  return (
    <div className="chat-container">
      <ModelSelector
        models={AVAILABLE_MODELS}
        selectedModel={selectedModel}
        onModelSelect={setSelectedModel}
        disabled={isLoading}
      />
      <div className="messages-container">
        {messages.map((message, index) => (
          <div
            key={index}
            className={`message ${message.role === 'user' ? 'user-message' : 'assistant-message'}`}
          >
            <div className="message-content">{message.content}</div>
            {message.role === 'assistant' && message.cost && (
              <div className="message-cost">
                {formatCost(message.cost)}
              </div>
            )}
          </div>
        ))}
        {/* Invisible element to scroll to */}
        <div ref={messagesEndRef} />
      </div>
      <div className="input-container">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && sendMessage(false)}
          placeholder="Type your message..."
          disabled={isLoading}
        />
        <div className="button-group">
          <button onClick={() => sendMessage(false)} disabled={isLoading}>
            {isLoading ? 'Sending...' : 'Send'}
          </button>
          <button 
            onClick={() => sendMessage(true)} 
            disabled={isLoading}
            className="stream-button"
          >
            Stream
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
