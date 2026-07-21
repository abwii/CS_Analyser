import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import AdminPage from './pages/AdminPage'
import { VetoPage } from './pages/VetoPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/veto/:sessionId" element={<VetoPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
