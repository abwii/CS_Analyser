import type { CanvasSize, Point2D, WorldBounds } from '../types'

export function projectToCanvas(point: Point2D, bounds: WorldBounds, size: CanvasSize): Point2D {
  const rangeX = bounds.maxX - bounds.minX
  const rangeY = bounds.maxY - bounds.minY

  const normalizedX = rangeX === 0 ? 0.5 : (point.x - bounds.minX) / rangeX
  const normalizedY = rangeY === 0 ? 0.5 : (bounds.maxY - point.y) / rangeY

  return {
    x: normalizedX * size.width,
    y: normalizedY * size.height,
  }
}
