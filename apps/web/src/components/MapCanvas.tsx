import { useEffect, useRef, useState } from 'react'
import { projectToCanvas } from '../lib/projection'
import type { Point2D, WorldBounds } from '../types'

interface MapCanvasProps {
  mapImageUrl: string
  bounds: WorldBounds
  points: Point2D[]
  width?: number
  height?: number
  pointColor?: string
  pointRadius?: number
}

export function MapCanvas({
  mapImageUrl,
  bounds,
  points,
  width = 800,
  height = 800,
  pointColor = '#ef4444',
  pointRadius = 4,
}: MapCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const [image, setImage] = useState<HTMLImageElement | null>(null)
  const [imageError, setImageError] = useState(false)

  useEffect(() => {
    setImage(null)
    setImageError(false)

    const img = new Image()
    img.onload = () => setImage(img)
    img.onerror = () => setImageError(true)
    img.src = mapImageUrl

    return () => {
      img.onload = null
      img.onerror = null
    }
  }, [mapImageUrl])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !image) {
      return
    }
    const context = canvas.getContext('2d')
    if (!context) {
      return
    }

    context.clearRect(0, 0, width, height)
    context.drawImage(image, 0, 0, width, height)

    context.fillStyle = pointColor
    for (const point of points) {
      const projected = projectToCanvas(point, bounds, { width, height })
      context.beginPath()
      context.arc(projected.x, projected.y, pointRadius, 0, Math.PI * 2)
      context.fill()
    }
  }, [image, points, bounds, width, height, pointColor, pointRadius])

  return (
    <div className="relative overflow-hidden rounded-lg border border-slate-200 bg-slate-100" style={{ width, height }}>
      {!image && !imageError && (
        <div className="absolute inset-0 flex items-center justify-center text-sm text-slate-400">Loading map</div>
      )}
      {imageError && (
        <div className="absolute inset-0 flex items-center justify-center text-sm text-red-500">Failed to load map</div>
      )}
      <canvas ref={canvasRef} width={width} height={height} className="h-full w-full" />
    </div>
  )
}
