import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { RadarChart } from './RadarChart'
import type { RadarPlayerData } from '../types'

const playerA: RadarPlayerData = {
  playerId: 'a',
  playerName: 'Alice',
  metrics: { kpr: 1.2, dpr: 0.5, kast: 80, impact: 1.8, adr: 90 },
}

const playerB: RadarPlayerData = {
  playerId: 'b',
  playerName: 'Bob',
  metrics: { kpr: 0.6, dpr: 0.9, kast: 55, impact: 0.9, adr: 60 },
}

describe('RadarChart', () => {
  it('renders five grid rings, five axis lines, and two distinct polygons', () => {
    const markup = renderToStaticMarkup(<RadarChart playerA={playerA} playerB={playerB} size={400} />)

    const polygons = markup.match(/<polygon[^>]*>/g) ?? []
    expect(polygons).toHaveLength(7)

    const lines = markup.match(/<line[^>]*>/g) ?? []
    expect(lines).toHaveLength(5)

    const points = polygons.map((polygon) => /points="([^"]*)"/.exec(polygon)?.[1])
    const playerPolygonPoints = points.slice(5)
    expect(playerPolygonPoints[0]).not.toBe(playerPolygonPoints[1])

    expect(markup).toContain('Alice')
    expect(markup).toContain('Bob')
    expect(markup).toContain('KPR')
    expect(markup).toContain('ADR')
  })

  it('renders identical polygons for identical metrics', () => {
    const markup = renderToStaticMarkup(<RadarChart playerA={playerA} playerB={playerA} size={400} />)
    const polygons = markup.match(/<polygon[^>]*>/g) ?? []
    const points = polygons.map((polygon) => /points="([^"]*)"/.exec(polygon)?.[1])
    const playerPolygonPoints = points.slice(5)
    expect(playerPolygonPoints[0]).toBe(playerPolygonPoints[1])
  })
})
