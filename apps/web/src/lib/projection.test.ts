import { describe, expect, it } from 'vitest'
import { projectToCanvas } from './projection'

const bounds = { minX: -100, maxX: 100, minY: -50, maxY: 50 }
const size = { width: 800, height: 400 }

describe('projectToCanvas', () => {
  it('maps the bottom left world corner to the bottom left pixel', () => {
    const result = projectToCanvas({ x: -100, y: -50 }, bounds, size)
    expect(result.x).toBe(0)
    expect(result.y).toBe(400)
  })

  it('maps the top right world corner to the top right pixel', () => {
    const result = projectToCanvas({ x: 100, y: 50 }, bounds, size)
    expect(result.x).toBe(800)
    expect(result.y).toBe(0)
  })

  it('maps the world center to the canvas center', () => {
    const result = projectToCanvas({ x: 0, y: 0 }, bounds, size)
    expect(result.x).toBe(400)
    expect(result.y).toBe(200)
  })

  it('flips the vertical axis so higher world y sits higher on screen', () => {
    const low = projectToCanvas({ x: 0, y: -50 }, bounds, size)
    const high = projectToCanvas({ x: 0, y: 50 }, bounds, size)
    expect(high.y).toBeLessThan(low.y)
  })

  it('does not flip the horizontal axis', () => {
    const left = projectToCanvas({ x: -100, y: 0 }, bounds, size)
    const right = projectToCanvas({ x: 100, y: 0 }, bounds, size)
    expect(right.x).toBeGreaterThan(left.x)
  })

  it('falls back to the midpoint when the world range is zero', () => {
    const degenerate = { minX: 10, maxX: 10, minY: 5, maxY: 5 }
    const result = projectToCanvas({ x: 10, y: 5 }, degenerate, size)
    expect(result.x).toBe(400)
    expect(result.y).toBe(200)
  })

  it('scales proportionally for an off center point', () => {
    const result = projectToCanvas({ x: -50, y: 25 }, bounds, size)
    expect(result.x).toBe(200)
    expect(result.y).toBe(100)
  })
})
