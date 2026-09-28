interface MetricCardProps {
  title: string
  value: string | number
  unit?: string
}

export default function MetricCard({ title, value, unit }: MetricCardProps) {
  return (
    <div className="metric-card">
      <div className="metric-title">{title}</div>
      <div className="metric-value">
        {value}
        {unit && <span className="metric-unit">{unit}</span>}
      </div>
    </div>
  )
}
