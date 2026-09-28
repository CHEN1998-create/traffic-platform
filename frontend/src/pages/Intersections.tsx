import { useEffect, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { api } from '../api/client'
import type { Intersection } from '../api/types'
import Chart from '../components/Chart'

export default function Intersections() {
  const [list, setList] = useState<Intersection[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .topIntersections(10)
      .then(setList)
      .catch((e: Error) => setError(e.message))
  }, [])

  // 横向柱状图，Y 轴倒序让最拥堵的路口显示在顶部
  const option: EChartsOption = {
    title: { text: '拥堵路口排行（Top10）' },
    tooltip: { trigger: 'axis' },
    grid: { left: 130 },
    xAxis: { type: 'value', name: '拥堵指数' },
    yAxis: { type: 'category', data: list.map((i) => i.intersectionId).reverse() },
    series: [
      {
        name: '拥堵指数',
        type: 'bar',
        data: list.map((i) => Math.round(i.congestionIndex * 10) / 10).reverse(),
      },
    ],
  }

  return (
    <div>
      <h2>路口排行</h2>
      {error ? (
        <div className="error">加载失败：{error}</div>
      ) : (
        <>
          <Chart option={option} height={360} />
          <table className="table">
            <thead>
              <tr>
                <th>路口</th>
                <th>总车流</th>
                <th>平均车速</th>
                <th>拥堵指数</th>
              </tr>
            </thead>
            <tbody>
              {list.map((i) => (
                <tr key={i.intersectionId}>
                  <td>{i.intersectionId}</td>
                  <td>{i.totalVehicles}</td>
                  <td>{i.avgSpeed.toFixed(1)}</td>
                  <td>{i.congestionIndex.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  )
}
