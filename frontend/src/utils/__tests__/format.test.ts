import { describe, it, expect } from 'vitest'
import { formatBytes, formatDate, truncate } from '../format'

describe('formatBytes', () => {
  it('should format 0 bytes as "0 B"', () => {
    expect(formatBytes(0)).toBe('0 B')
  })

  it('should format bytes under 1024 as B', () => {
    expect(formatBytes(512)).toBe('512 B')
    expect(formatBytes(1023)).toBe('1023 B')
  })

  it('should format kilobytes', () => {
    expect(formatBytes(1024)).toBe('1.0 KB')
    expect(formatBytes(1536)).toBe('1.5 KB')
  })

  it('should format megabytes', () => {
    expect(formatBytes(1024 * 1024)).toBe('1.0 MB')
    expect(formatBytes(1024 * 1024 * 5.25)).toBe('5.3 MB')
  })

  it('should format gigabytes', () => {
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1.0 GB')
  })

  it('should handle negative bytes', () => {
    expect(formatBytes(-1536)).toBe('-1.5 KB')
  })
})

describe('formatDate', () => {
  it('should format a Date object as YYYY-MM-DD HH:mm:ss', () => {
    const d = new Date(2024, 0, 15, 9, 5, 3) // month is 0-indexed
    expect(formatDate(d)).toBe('2024-01-15 09:05:03')
  })

  it('should format a date string', () => {
    expect(formatDate('2024-06-01T12:30:00')).toBe('2024-06-01 12:30:00')
  })

  it('should pad single-digit values with leading zero', () => {
    const d = new Date(2024, 0, 1, 0, 0, 0)
    expect(formatDate(d)).toBe('2024-01-01 00:00:00')
  })

  it('should return "-" for invalid dates', () => {
    expect(formatDate('invalid-date')).toBe('-')
  })
})

describe('truncate', () => {
  it('should return the original string if shorter than maxLen', () => {
    expect(truncate('hello', 10)).toBe('hello')
  })

  it('should truncate strings longer than maxLen with ellipsis', () => {
    expect(truncate('hello world', 8)).toBe('hello w…')
  })

  it('should return empty string for maxLen <= 0', () => {
    expect(truncate('hello', 0)).toBe('')
    expect(truncate('hello', -1)).toBe('')
  })

  it('should return only ellipsis for maxLen = 1', () => {
    expect(truncate('hello', 1)).toBe('…')
  })

  it('should handle null/undefined input gracefully', () => {
    expect(truncate(null as unknown as string, 5)).toBe('')
    expect(truncate(undefined as unknown as string, 5)).toBe('')
  })
})
