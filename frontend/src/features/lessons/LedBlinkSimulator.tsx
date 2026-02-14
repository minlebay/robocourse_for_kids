import { useCallback, useEffect, useRef, useState } from 'react'

const DEFAULT_CODE = `from machine import Pin
import time
time.sleep(0.1) # Wait for USB to become ready

led = Pin("LED", Pin.OUT)
while True:
    led.value(1)
    time.sleep(0.5)
    led.value(0)
    time.sleep(0.5)
`

type Cmd = { type: 'on' } | { type: 'off' } | { type: 'sleep'; sec: number }

function parseCode(code: string): Cmd[] {
  const lines = code.split(/\r?\n/)
  const commands: Cmd[] = []
  let inWhile = false
  let whileIndent = -1

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    const indent = line.search(/\S/)
    if (/while\s+True\s*:/.test(trimmed)) {
      inWhile = true
      whileIndent = indent >= 0 ? indent : 0
      continue
    }
    if (inWhile && indent <= whileIndent && trimmed) {
      inWhile = false
    }
    if (!inWhile && whileIndent >= 0) continue

    if (/led\.value\s*\(\s*1\s*\)|led\.on\s*\(\s*\)/.test(trimmed)) {
      commands.push({ type: 'on' })
    } else if (/led\.value\s*\(\s*0\s*\)|led\.off\s*\(\s*\)/.test(trimmed)) {
      commands.push({ type: 'off' })
    } else {
      const m = trimmed.match(/time\.sleep\s*\(\s*([\d.]+)\s*\)/)
      if (m) {
        const sec = Math.min(5, Math.max(0.05, parseFloat(m[1]) || 0.5))
        commands.push({ type: 'sleep', sec })
      }
    }
  }

  return commands.length ? commands : [
    { type: 'on' },
    { type: 'sleep', sec: 0.5 },
    { type: 'off' },
    { type: 'sleep', sec: 0.5 },
  ]
}

export function LedBlinkSimulator() {
  const [code, setCode] = useState(DEFAULT_CODE)
  const [ledOn, setLedOn] = useState(false)
  const [running, setRunning] = useState(false)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const loopRef = useRef(false)

  const stop = useCallback(() => {
    loopRef.current = false
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    setRunning(false)
  }, [])

  const run = useCallback(() => {
    stop()
    const commands = parseCode(code)
    if (!commands.length) return
    setRunning(true)
    loopRef.current = true

    let iter = 0
    const maxIter = 20

    const runNext = (idx: number) => {
      if (!loopRef.current) {
        setRunning(false)
        return
      }
      if (idx >= commands.length) {
        iter++
        if (iter >= maxIter) {
          setRunning(false)
          return
        }
        runNext(0)
        return
      }
      const cmd = commands[idx]
      if (cmd.type === 'on') {
        setLedOn(true)
        runNext(idx + 1)
      } else if (cmd.type === 'off') {
        setLedOn(false)
        runNext(idx + 1)
      } else {
        timeoutRef.current = setTimeout(() => {
          timeoutRef.current = null
          runNext(idx + 1)
        }, cmd.sec * 1000)
      }
    }
    runNext(0)
  }, [code, stop])

  useEffect(() => () => stop(), [stop])

  return (
    <div className="simulator simulator-led">
      <div className="simulator-header">
        <span className="simulator-title">Тренажёр: мигающий светодиод</span>
        <div className="simulator-actions">
          <button
            type="button"
            className="simulator-btn run"
            onClick={run}
            disabled={running}
          >
            Запуск
          </button>
          <button
            type="button"
            className="simulator-btn stop"
            onClick={stop}
            disabled={!running}
          >
            Стоп
          </button>
        </div>
      </div>
      <div className="simulator-body">
        <div className="simulator-board" aria-label="Плата Pico">
          <div className="simulator-board-inner">
            <span className="simulator-board-title">Pico</span>
            <div className="simulator-led-wrap">
              <div
                className={`simulator-led ${ledOn ? 'on' : ''}`}
                title={ledOn ? 'Включён' : 'Выключен'}
                aria-hidden
              />
              <span className="simulator-led-label">LED</span>
            </div>
          </div>
        </div>
        <div className="simulator-editor-wrap">
          <label className="simulator-editor-label">Код (MicroPython)</label>
          <textarea
            className="simulator-editor"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            spellCheck={false}
            rows={14}
            disabled={running}
          />
        </div>
      </div>
      <p className="simulator-hint">
        Для встроенного светодиода: <code>led = Pin("LED", Pin.OUT)</code>. Поддерживаются команды: <code>led.value(1)</code> / <code>led.value(0)</code>, <code>time.sleep(секунды)</code>. Цикл повторяется несколько раз, затем останавливается.
      </p>
    </div>
  )
}
