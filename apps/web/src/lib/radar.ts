import type { PlayerMetrics, RadarAxisConfig, RadarPoint } from '../types'

export const DEFAULT_RADAR_AXES: RadarAxisConfig[] = [
  { key: 'kpr', label: 'KPR', min: 0, max: 1.5 },
  { key: 'dpr', label: 'DPR', min: 0, max: 1.5 },
  { key: 'kast', label: 'KAST', min: 0, max: 100 },
  { key: 'impact', label: 'Impact', min: 0, max: 3 },
  { key: 'adr', label: 'ADR', min: 0, max: 120 },
]

export function axisAngle(index: number, axisCount: number): number {
  return (Math.PI * 2 * index) / axisCount - Math.PI / 2
}

export function normalize(value: number, min: number, max: number): number {
  if (max === min) {
    return 0
  }
  const clamped = Math.min(Math.max(value, min), max)
  return (clamped - min) / (max - min)
}

export function radarPoint(
  value: number,
  axis: RadarAxisConfig,
  index: number,
  axisCount: number,
  center: RadarPoint,
  radius: number,
): RadarPoint {
  const ratio = normalize(value, axis.min, axis.max)
  const angle = axisAngle(index, axisCount)
  return {
    x: center.x + ratio * radius * Math.cos(angle),
    y: center.y + ratio * radius * Math.sin(angle),
  }
}

export function buildPolygonPoints(
  metrics: PlayerMetrics,
  axes: RadarAxisConfig[],
  center: RadarPoint,
  radius: number,
): RadarPoint[] {
  return axes.map((axis, index) => radarPoint(metrics[axis.key], axis, index, axes.length, center, radius))
}

export function buildGridRing(
  ratio: number,
  axes: RadarAxisConfig[],
  center: RadarPoint,
  radius: number,
): RadarPoint[] {
  return axes.map((_, index) => {
    const angle = axisAngle(index, axes.length)
    return {
      x: center.x + ratio * radius * Math.cos(angle),
      y: center.y + ratio * radius * Math.sin(angle),
    }
  })
}

export function axisLabelPoint(
  index: number,
  axisCount: number,
  center: RadarPoint,
  radius: number,
  ratio = 1.18,
): RadarPoint {
  const angle = axisAngle(index, axisCount)
  return {
    x: center.x + ratio * radius * Math.cos(angle),
    y: center.y + ratio * radius * Math.sin(angle),
  }
}

export function pointsToAttr(points: RadarPoint[]): string {
  return points.map((point) => `${point.x},${point.y}`).join(' ')
}
