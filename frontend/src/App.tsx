import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import Overview from './pages/Overview'
import Trend from './pages/Trend'
import Intersections from './pages/Intersections'
import Alerts from './pages/Alerts'
import Imports from './pages/Imports'
import Operations from './pages/Operations'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Overview />} />
          <Route path="/dashboard/trend" element={<Trend />} />
          <Route path="/dashboard/intersections" element={<Intersections />} />
          <Route path="/alerts" element={<Alerts />} />
          <Route path="/imports" element={<Imports />} />
          <Route path="/operations" element={<Operations />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
