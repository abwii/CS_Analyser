import { DEFAULT_RADAR_AXES, axisLabelPoint, buildGridRing, buildPolygonPoints, pointsToAttr } from '../lib/radar'
import type { RadarAxisConfig, RadarPlayerData } from '../types'

const GRID_RATIOS = [0.2, 0.4, 0.6, 0.8, 1]

interface RadarChartProps {
  axes?: RadarAxisConfig[]
  playerA: RadarPlayerData
  playerB: RadarPlayerData
  size?: number
}

export function RadarChart({ axes = DEFAULT_RADAR_AXES, playerA, playerB, size = 400 }: RadarChartProps) {
  const center = { x: size / 2, y: size / 2 }
  const radius = size / 2 - size * 0.18

  const polygonA = buildPolygonPoints(playerA.metrics, axes, center, radius)
  const polygonB = buildPolygonPoints(playerB.metrics, axes, center, radius)
  const outerRing = buildGridRing(1, axes, center, radius)

  return (
    <div className="inline-flex flex-col items-center gap-4 rounded-lg border border-slate-200 bg-white p-4">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {GRID_RATIOS.map((ratio) => (
          <polygon
            key={ratio}
            points={pointsToAttr(buildGridRing(ratio, axes, center, radius))}
            fill="none"
            stroke="#e2e8f0"
            strokeWidth={1}
          />
        ))}

        {axes.map((axis, index) => (
          <line
            key={axis.key}
            x1={center.x}
            y1={center.y}
            x2={outerRing[index].x}
            y2={outerRing[index].y}
            stroke="#e2e8f0"
            strokeWidth={1}
          />
        ))}

        <polygon points={pointsToAttr(polygonA)} fill="#3b82f6" fillOpacity={0.35} stroke="#3b82f6" strokeWidth={2} />
        <polygon points={pointsToAttr(polygonB)} fill="#ef4444" fillOpacity={0.35} stroke="#ef4444" strokeWidth={2} />

        {axes.map((axis, index) => {
          const label = axisLabelPoint(index, axes.length, center, radius)
          return (
            <text
              key={axis.key}
              x={label.x}
              y={label.y}
              textAnchor="middle"
              dominantBaseline="middle"
              className="fill-slate-500 text-xs"
            >
              {axis.label}
            </text>
          )
        })}
      </svg>

      <div className="flex gap-6 text-sm text-slate-700">
        <span className="flex items-center gap-2">
          <span className="h-2.5 w-2.5 rounded-full bg-blue-500" />
          {playerA.playerName}
        </span>
        <span className="flex items-center gap-2">
          <span className="h-2.5 w-2.5 rounded-full bg-red-500" />
          {playerB.playerName}
        </span>
      </div>
    </div>
  )
}
