import { render, screen } from '@testing-library/react'
import App from './App'

test('uygulama başlığı görüntülenir', () => {
  render(<App />)
  expect(screen.getByText(/PUYAD Kalkanı/i)).toBeInTheDocument()
})
