import { useCallback, useEffect, useMemo, useState } from 'react'
import { Circle, Group, Image as KonvaImage, Layer, Line, Rect, Stage, Text, Wedge } from 'react-konva'
import { api } from '../api/client'
import type { AoETemplate, AoETemplateShape, GameMap, Token } from '../api/types'

type RevealedBox = {
  x: number
  y: number
  width: number
  height: number
}

type FloatingText = {
  id: string
  x: number
  y: number
  text: string
  color: string
  opacity: number
}

type Props = {
  map: GameMap
  tokens: Token[]
  canMove: (token: Token) => boolean
  selectedTokenId: number | null
  onSelect: (tokenId: number | null) => void
  onMove: (id: number, x: number, y: number) => void
  isDM?: boolean
  onUpdateMap?: (updated: GameMap) => void
}

function snap(value: number, gridSize: number): number {
  return Math.round((value - gridSize / 2) / gridSize) * gridSize + gridSize / 2
}

function useMapImage(url: string): HTMLImageElement | null {
  const [image, setImage] = useState<HTMLImageElement | null>(null)

  useEffect(() => {
    if (!url) {
      setImage(null)
      return
    }
    const element = new window.Image()
    element.crossOrigin = 'anonymous'
    element.src = url
    element.onload = () => setImage(element)
    return () => {
      element.onload = null
    }
  }, [url])

  return image
}

function healthColor(ratio: number): string {
  if (ratio > 0.5) return '#3ddc84'
  if (ratio > 0.25) return '#e9b949'
  return '#e4572e'
}

export function MapCanvas({
  map,
  tokens,
  canMove,
  selectedTokenId,
  onSelect,
  onMove,
  isDM = false,
  onUpdateMap,
}: Props) {
  const background = useMapImage(map.image_url)

  // Map modes: 'select' | 'ruler' | 'fog_reveal' | 'fog_hide' | 'aoe'
  const [mode, setMode] = useState<'select' | 'ruler' | 'fog_reveal' | 'fog_hide' | 'aoe'>('select')

  // Ruler state
  const [rulerStart, setRulerStart] = useState<{ x: number; y: number } | null>(null)
  const [rulerEnd, setRulerEnd] = useState<{ x: number; y: number } | null>(null)

  // Fog of War state
  const revealedBoxes: RevealedBox[] = useMemo(() => {
    if (!map.fog_revealed) return []
    try {
      return JSON.parse(map.fog_revealed) as RevealedBox[]
    } catch {
      return []
    }
  }, [map.fog_revealed])

  const [dragBox, setDragBox] = useState<{ x: number; y: number; width: number; height: number } | null>(null)
  const [boxStart, setBoxStart] = useState<{ x: number; y: number } | null>(null)

  // AoE templates state
  const [aoeTemplates, setAoeTemplates] = useState<AoETemplate[]>([])
  const [aoeShape, setAoeShape] = useState<AoETemplateShape>('circle')
  const [aoeSizeFt, setAoeSizeFt] = useState(20)
  const [aoeColor, setAoeColor] = useState('#e74c3c')

  // Floating combat text
  const [floatingTexts] = useState<FloatingText[]>([])

  // Load AoE templates
  const refreshAoE = useCallback(() => {
    api
      .listAoETemplates(map.id)
      .then((res) => setAoeTemplates(Array.isArray(res) ? res : []))
      .catch(() => setAoeTemplates([]))
  }, [map.id])

  useEffect(() => {
    refreshAoE()
  }, [refreshAoE])

  // Grid background lines
  const gridLines = useMemo(() => {
    const lines: number[][] = []
    const gridSize = map.grid_size || 50
    const width = map.width || 1920
    const height = map.height || 1080
    for (let x = gridSize; x < width; x += gridSize) {
      lines.push([x, 0, x, height])
    }
    for (let y = gridSize; y < height; y += gridSize) {
      lines.push([0, y, width, y])
    }
    return lines
  }, [map.grid_size, map.width, map.height])

  // Distance calculation in feet
  const rulerDistance = useMemo(() => {
    if (!rulerStart || !rulerEnd) return null
    const gridSize = map.grid_size || 50
    const dx = rulerEnd.x - rulerStart.x
    const dy = rulerEnd.y - rulerStart.y
    const pixelDist = Math.hypot(dx, dy)
    const squares = Math.max(1, Math.round(pixelDist / gridSize))
    const feet = squares * 5
    return { feet, squares, midX: (rulerStart.x + rulerEnd.x) / 2, midY: (rulerStart.y + rulerEnd.y) / 2 }
  }, [rulerStart, rulerEnd, map.grid_size])

  // Stage Mouse Handlers
  const handleMouseDown = (e: { target: { getStage: () => unknown }; evt: MouseEvent }) => {
    const stage = (e.target as unknown as { getStage: () => { getPointerPosition: () => { x: number; y: number } | null } }).getStage()
    const pos = stage?.getPointerPosition()
    if (!pos) return

    if (mode === 'ruler') {
      setRulerStart(pos)
      setRulerEnd(pos)
      return
    }

    if (mode === 'fog_reveal' || mode === 'fog_hide') {
      setBoxStart(pos)
      setDragBox({ x: pos.x, y: pos.y, width: 0, height: 0 })
      return
    }

    if (mode === 'aoe') {
      // Place new AoE template
      api
        .createAoETemplate(map.id, {
          name: `${aoeSizeFt}ft ${aoeShape}`,
          shape: aoeShape,
          size_ft: aoeSizeFt,
          color: aoeColor,
          x: Math.round(pos.x),
          y: Math.round(pos.y),
        })
        .then((created) => {
          setAoeTemplates((curr) => [...curr, created])
          setMode('select')
        })
        .catch(() => {})
      return
    }

    if (e.target === (stage as unknown)) {
      onSelect(null)
    }
  }

  const handleMouseMove = (e: { target: { getStage: () => unknown } }) => {
    const stage = (e.target as unknown as { getStage: () => { getPointerPosition: () => { x: number; y: number } | null } }).getStage()
    const pos = stage?.getPointerPosition()
    if (!pos) return

    if (mode === 'ruler' && rulerStart) {
      setRulerEnd(pos)
      return
    }

    if ((mode === 'fog_reveal' || mode === 'fog_hide') && boxStart) {
      const x = Math.min(boxStart.x, pos.x)
      const y = Math.min(boxStart.y, pos.y)
      const width = Math.abs(pos.x - boxStart.x)
      const height = Math.abs(pos.y - boxStart.y)
      setDragBox({ x, y, width, height })
    }
  }

  const handleMouseUp = async () => {
    if (mode === 'ruler') {
      // Keep ruler visible until clicked or mode switched
      return
    }

    if ((mode === 'fog_reveal' || mode === 'fog_hide') && dragBox && dragBox.width > 5 && dragBox.height > 5) {
      let updatedBoxes: RevealedBox[] = [...revealedBoxes]
      if (mode === 'fog_reveal') {
        updatedBoxes.push(dragBox)
      } else {
        // Hide: filter out or subtract (ponytail: keep simple box clear)
        updatedBoxes = updatedBoxes.filter(
          (b) =>
            !(
              b.x >= dragBox.x &&
              b.y >= dragBox.y &&
              b.x + b.width <= dragBox.x + dragBox.width &&
              b.y + b.height <= dragBox.y + dragBox.height
            ),
        )
      }
      setBoxStart(null)
      setDragBox(null)
      try {
        const updated = await api.updateMapFog(map.id, {
          fog_enabled: map.fog_enabled ?? true,
          fog_revealed: JSON.stringify(updatedBoxes),
        })
        onUpdateMap?.(updated)
      } catch {
        // Ignored
      }
    } else {
      setBoxStart(null)
      setDragBox(null)
    }
  }

  const handleToggleFog = async () => {
    try {
      const nextEnabled = !map.fog_enabled
      const updated = await api.updateMapFog(map.id, {
        fog_enabled: nextEnabled,
        fog_revealed: map.fog_revealed || '[]',
      })
      onUpdateMap?.(updated)
    } catch {
      // Ignored
    }
  }

  const handleClearFog = async () => {
    if (!confirm('Reveal the entire map?')) return
    try {
      const updated = await api.updateMapFog(map.id, {
        fog_enabled: false,
        fog_revealed: '[]',
      })
      onUpdateMap?.(updated)
    } catch {
      // Ignored
    }
  }

  const handleDeleteAoE = async (id: number) => {
    try {
      await api.deleteAoETemplate(id)
      setAoeTemplates((curr) => curr.filter((t) => t.id !== id))
    } catch {
      // Ignored
    }
  }

  return (
    <div className="map-container">
      {/* Interactive Map Toolbar */}
      <div className="map-toolbar">
        <div className="toolbar-group">
          <button
            className={`toolbar-btn ${mode === 'select' ? 'active' : ''}`}
            onClick={() => {
              setMode('select')
              setRulerStart(null)
              setRulerEnd(null)
            }}
            title="Select & Move Tokens"
          >
            ✋ Select
          </button>
          <button
            className={`toolbar-btn ${mode === 'ruler' ? 'active' : ''}`}
            onClick={() => {
              setMode('ruler')
              setRulerStart(null)
              setRulerEnd(null)
            }}
            title="Measure Distance in Feet"
          >
            📏 Ruler
          </button>
          <button
            className={`toolbar-btn ${mode === 'aoe' ? 'active' : ''}`}
            onClick={() => setMode('aoe')}
            title="Place AoE Spell Template"
          >
            🔥 AoE Spell
          </button>
        </div>

        {mode === 'aoe' && (
          <div className="toolbar-subgroup">
            <select
              value={aoeShape}
              onChange={(e) => setAoeShape(e.target.value as AoETemplateShape)}
            >
              <option value="circle">Circle (Sphere/Cylinder)</option>
              <option value="cone">Cone</option>
              <option value="line">Line</option>
              <option value="cube">Cube / Square</option>
            </select>
            <input
              type="number"
              min={5}
              max={150}
              step={5}
              value={aoeSizeFt}
              onChange={(e) => setAoeSizeFt(parseInt(e.target.value, 10) || 20)}
              style={{ width: '60px' }}
            />
            <span>ft</span>
            <input
              type="color"
              value={aoeColor}
              onChange={(e) => setAoeColor(e.target.value)}
              title="Template Color"
            />
            <span className="hint">Click on map to place</span>
          </div>
        )}

        {isDM && (
          <div className="toolbar-group">
            <button
              className={`toolbar-btn ${map.fog_enabled ? 'active-warning' : ''}`}
              onClick={handleToggleFog}
              title="Toggle Fog of War on/off"
            >
              🌫️ Fog: {map.fog_enabled ? 'ON' : 'OFF'}
            </button>
            {map.fog_enabled && (
              <>
                <button
                  className={`toolbar-btn ${mode === 'fog_reveal' ? 'active' : ''}`}
                  onClick={() => setMode('fog_reveal')}
                  title="Drag rectangle to reveal area to players"
                >
                  👁️ Reveal Box
                </button>
                <button
                  className={`toolbar-btn ${mode === 'fog_hide' ? 'active' : ''}`}
                  onClick={() => setMode('fog_hide')}
                  title="Drag rectangle to hide area"
                >
                  🔒 Hide Box
                </button>
                <button
                  className="toolbar-btn text-danger"
                  onClick={handleClearFog}
                  title="Reveal whole map"
                >
                  Clear Fog
                </button>
              </>
            )}
          </div>
        )}
      </div>

      <Stage
        width={map.width}
        height={map.height}
        className="stage"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
      >
        {/* Layer 1: Background & Grid */}
        <Layer listening={false}>
          <Rect width={map.width} height={map.height} fill="#1b1d23" />
          {background && <KonvaImage image={background} width={map.width} height={map.height} />}
          {gridLines.map((points, index) => (
            <Line key={index} points={points} stroke="#ffffff" strokeWidth={1} opacity={0.08} />
          ))}
        </Layer>

        {/* Layer 2: AoE Spell Templates */}
        <Layer>
          {(aoeTemplates || []).map((t) => {
            const pixelRadius = (t.size_ft / 5) * map.grid_size
            return (
              <Group
                key={t.id}
                x={t.x}
                y={t.y}
                draggable={isDM}
                onClick={() => {
                  if (isDM && confirm(`Delete ${t.name}?`)) handleDeleteAoE(t.id)
                }}
              >
                {t.shape === 'circle' && (
                  <Circle
                    radius={pixelRadius}
                    fill={t.color}
                    opacity={0.35}
                    stroke={t.color}
                    strokeWidth={2}
                  />
                )}
                {t.shape === 'cube' && (
                  <Rect
                    x={-pixelRadius / 2}
                    y={-pixelRadius / 2}
                    width={pixelRadius}
                    height={pixelRadius}
                    fill={t.color}
                    opacity={0.35}
                    stroke={t.color}
                    strokeWidth={2}
                  />
                )}
                {t.shape === 'cone' && (
                  <Wedge
                    radius={pixelRadius}
                    angle={60}
                    rotation={-30}
                    fill={t.color}
                    opacity={0.35}
                    stroke={t.color}
                    strokeWidth={2}
                  />
                )}
                {t.shape === 'line' && (
                  <Rect
                    x={0}
                    y={-map.grid_size / 2}
                    width={pixelRadius}
                    height={map.grid_size}
                    fill={t.color}
                    opacity={0.35}
                    stroke={t.color}
                    strokeWidth={2}
                  />
                )}
                <Text
                  text={t.name}
                  fontSize={11}
                  fill="#ffffff"
                  fontStyle="bold"
                  align="center"
                  offsetX={50}
                  width={100}
                  listening={false}
                />
              </Group>
            )
          })}
        </Layer>

        {/* Layer 3: Tokens */}
        <Layer>
          {(tokens || []).map((token) => {
            const radius = (token.size * (map.grid_size || 50)) / 2
            const movable = canMove(token)
            const targeted = token.id === selectedTokenId
            const dead = token.hp_max > 0 && token.hp_current === 0
            const ratio = token.hp_max > 0 ? token.hp_current / token.hp_max : 1

            return (
              <Group
                key={token.id}
                x={token.x}
                y={token.y}
                draggable={movable && mode === 'select'}
                opacity={dead ? 0.45 : 1}
                onClick={() => onSelect(token.id)}
                onTap={() => onSelect(token.id)}
                onDragEnd={(event) =>
                  onMove(
                    token.id,
                    snap(event.target.x(), map.grid_size),
                    snap(event.target.y(), map.grid_size),
                  )
                }
              >
                {targeted && (
                  <Circle radius={radius + 5} stroke="#6c8cff" strokeWidth={2} dash={[6, 4]} listening={false} />
                )}

                <Circle
                  radius={radius}
                  fill={token.color}
                  stroke={movable ? '#f4f4f5' : '#0d0f13'}
                  strokeWidth={movable ? 3 : 2}
                  opacity={movable ? 1 : 0.85}
                />

                {dead && (
                  <Text
                    text="X"
                    fontSize={radius}
                    fontStyle="bold"
                    fill="#0d0f13"
                    width={radius * 2}
                    align="center"
                    offsetX={radius}
                    offsetY={radius / 2}
                    listening={false}
                  />
                )}

                {token.hp_max > 0 && (
                  <Group y={-radius - 10} listening={false}>
                    <Rect
                      x={-radius}
                      width={radius * 2}
                      height={5}
                      fill="#0d0f13"
                      cornerRadius={2}
                      opacity={0.8}
                    />
                    <Rect
                      x={-radius}
                      width={Math.max(0, radius * 2 * ratio)}
                      height={5}
                      fill={healthColor(ratio)}
                      cornerRadius={2}
                    />
                  </Group>
                )}

                {(token.conditions || []).length > 0 && (
                  <Text
                    text={(token.conditions || []).map((c) => c.slice(0, 4)).join(' ')}
                    fontSize={9}
                    fill="#e9b949"
                    width={radius * 4}
                    align="center"
                    offsetX={radius * 2}
                    y={-radius - 22}
                    listening={false}
                  />
                )}

                <Text
                  text={token.name}
                  fontSize={12}
                  fill="#f4f4f5"
                  width={radius * 4}
                  align="center"
                  offsetX={radius * 2}
                  y={radius + 4}
                  listening={false}
                />
              </Group>
            )
          })}
        </Layer>

        {/* Layer 4: Fog of War */}
        {map.fog_enabled && (
          <Layer listening={false}>
            {/* Shroud background covering entire map */}
            <Rect
              width={map.width}
              height={map.height}
              fill="#08090b"
              opacity={isDM ? 0.55 : 1.0}
            />

            {/* Revealed cutouts */}
            {(revealedBoxes || []).map((box, index) => (
              <Group key={index}>
                <Rect
                  x={box.x}
                  y={box.y}
                  width={box.width}
                  height={box.height}
                  fill="#000000"
                  globalCompositeOperation="destination-out"
                />
                {isDM && (
                  <Rect
                    x={box.x}
                    y={box.y}
                    width={box.width}
                    height={box.height}
                    stroke="#f1c40f"
                    strokeWidth={1}
                    dash={[4, 4]}
                    opacity={0.7}
                  />
                )}
              </Group>
            ))}
          </Layer>
        )}

        {/* Layer 5: Drag Box / Ruler / Floating Overlays */}
        <Layer listening={false}>
          {dragBox && (
            <Rect
              x={dragBox.x}
              y={dragBox.y}
              width={dragBox.width}
              height={dragBox.height}
              fill={mode === 'fog_reveal' ? 'rgba(46, 204, 113, 0.25)' : 'rgba(231, 76, 60, 0.25)'}
              stroke={mode === 'fog_reveal' ? '#2ecc71' : '#e74c3c'}
              strokeWidth={2}
              dash={[6, 3]}
            />
          )}

          {rulerStart && rulerEnd && rulerDistance && (
            <Group>
              <Line
                points={[rulerStart.x, rulerStart.y, rulerEnd.x, rulerEnd.y]}
                stroke="#6c8cff"
                strokeWidth={3}
                dash={[8, 4]}
              />
              <Circle x={rulerStart.x} y={rulerStart.y} radius={5} fill="#6c8cff" />
              <Circle x={rulerEnd.x} y={rulerEnd.y} radius={5} fill="#6c8cff" />

              {/* Distance badge */}
              <Group x={rulerDistance.midX - 50} y={rulerDistance.midY - 20}>
                <Rect
                  width={100}
                  height={24}
                  fill="#0d0f13"
                  stroke="#6c8cff"
                  strokeWidth={1}
                  cornerRadius={4}
                  opacity={0.9}
                />
                <Text
                  text={`${rulerDistance.feet} ft (${rulerDistance.squares} sq)`}
                  fontSize={11}
                  fontStyle="bold"
                  fill="#f4f4f5"
                  width={100}
                  height={24}
                  align="center"
                  verticalAlign="middle"
                  padding={4}
                />
              </Group>
            </Group>
          )}

          {/* Floating Combat Numbers */}
          {(floatingTexts || []).map((ft) => (
            <Text
              key={ft.id}
              x={ft.x - 30}
              y={ft.y - 25}
              text={ft.text}
              fontSize={14}
              fontStyle="bold"
              fill={ft.color}
              opacity={ft.opacity}
              width={60}
              align="center"
            />
          ))}
        </Layer>
      </Stage>
    </div>
  )
}
