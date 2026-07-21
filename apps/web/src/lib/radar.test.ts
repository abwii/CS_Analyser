import { describe, expect, it } from 'vitest'
import { axisAngle, axisLabelPoint, buildGridRing, buildPolygonPoints, normalize, pointsToAttr, radarPoint } from './radar'
import type { PlayerMetrics, RadarAxisConfig } from '../types'

const axes: RadarAxisConfig[] = [
  { key: 'kpr', label: 'KPR', min: 0, max: 1 },
  { key: 'dpr', label: 'DPR', min: 0, max: 1 },
  { key: 'kast', label: 'KAST', min: 0, max: 100 },
  { key: 'impact', label: 'Impact', min: 0, max: 2 },
  { key: 'adr', label: 'ADR', min: 0, max: 100 },
]

const center = { x: 100, y: 100 }
const radius = 100

describe('normalize', () => {
  it('maps min to 0 and max to 1', () => {
    expect(normalize(0, 0, 100)).toBe(0)
    expect(normalize(100, 0, 100)).toBe(1)
  })

  it('clamps values outside the range', () => {
    expect(normalize(-10, 0, 100)).toBe(0)
    expect(normalize(150, 0, 100)).toBe(1)
  })

  it('returns 0 when min equals max', () => {
    expect(normalize(5, 5, 5)).toBe(0)
  })
})

describe('axisAngle', () => {
  it('places the first axis at the top', () => {
    expect(axisAngle(0, 5)).toBeCloseTo(-Math.PI / 2)
  })

  it('spaces axes evenly around the circle', () => {
    const step = axisAngle(1, 5) - axisAngle(0, 5)
    expect(step).toBeCloseTo((Math.PI * 2) / 5)
  })
})

describe('radarPoint', () => {
  it('places a zero value at the center', () => {
    const point = radarPoint(0, axes[0], 0, axes.length, center, radius)
    expect(point.x).toBeCloseTo(center.x)
    expect(point.y).toBeCloseTo(center.y)
  })

  it('places a max value at full radius along the axis angle', () => {
    const point = radarPoint(1, axes[0], 0, axes.length, center, radius)
    expect(point.x).toBeCloseTo(center.x + radius * Math.cos(axisAngle(0, axes.length)))
    expect(point.y).toBeCloseTo(center.y + radius * Math.sin(axisAngle(0, axes.length)))
  })
})

describe('buildPolygonPoints', () => {
  const metrics: PlayerMetrics = { kpr: 1, dpr: 0, kast: 50, impact: 1, adr: 50 }

  it('produces one point per axis', () => {
    const points = buildPolygonPoints(metrics, axes, center, radius)
    expect(points).toHaveLength(axes.length)
  })

  it('collapses a zero metric onto the center', () => {
    const points = buildPolygonPoints(metrics, axes, center, radius)
    expect(points[1].x).toBeCloseTo(center.x)
    expect(points[1].y).toBeCloseTo(center.y)
  })
})

describe('buildGridRing', () => {
  it('scales every vertex by the given ratio', () => {
    const full = buildGridRing(1, axes, center, radius)
    const half = buildGridRing(0.5, axes, center, radius)
    full.forEach((point, index) => {
      const dxFull = point.x - center.x
      const dyFull = point.y - center.y
      const dxHalf = half[index].x - center.x
      const dyHalf = half[index].y - center.y
      expect(dxHalf).toBeCloseTo(dxFull / 2)
      expect(dyHalf).toBeCloseTo(dyFull / 2)
    })
  })
})

describe('axisLabelPoint', () => {
  it('sits further out than the grid ring', () => {
    const ring = buildGridRing(1, axes, center, radius)[0]
    const label = axisLabelPoint(0, axes.length, center, radius)
    const ringDistance = Math.hypot(ring.x - center.x, ring.y - center.y)
    const labelDistance = Math.hypot(label.x - center.x, label.y - center.y)
    expect(labelDistance).toBeGreaterThan(ringDistance)
  })
})

describe('pointsToAttr', () => {
  it('formats points as a space separated svg points string', () => {
    expect(pointsToAttr([{ x: 1, y: 2 }, { x: 3, y: 4 }])).toBe('1,2 3,4')
  })

  it('returns an empty string for no points', () => {
    expect(pointsToAttr([])).toBe('')
  })
})
