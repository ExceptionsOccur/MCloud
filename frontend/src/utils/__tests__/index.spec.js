import { describe, it, expect } from 'vitest'
import {
  formatDate,
  formatTime,
  toRFC3339,
  toLocalPickerValue,
  statusOptions,
  envTypeOptions,
  assetTypeOptions,
  filterCloudRegions
} from '../index.js'

describe('utils 冒烟', () => {
  it('formatDate 输出 YYYY-MM-DD', () => {
    expect(formatDate(new Date(2026, 9, 10))).toBe('2026-10-10')
  })

  it('formatDate 非法输入返回空串', () => {
    expect(formatDate('not-a-date')).toBe('')
    expect(formatDate(new Date('bad'))).toBe('')
  })

  it('formatTime 空值返回占位符', () => {
    expect(formatTime(null)).toBe('-')
    expect(formatTime(undefined)).toBe('-')
    expect(formatTime('')).toBe('-')
  })

  it('toRFC3339 本地无时区值可往返', () => {
    const iso = toRFC3339('2026-10-10T08:30:00')
    expect(iso).toBeTruthy()
    expect(new Date(iso).toISOString()).toBe(iso)
  })

  it('toLocalPickerValue 空值返回空串', () => {
    expect(toLocalPickerValue('')).toBe('')
    expect(toLocalPickerValue(null)).toBe('')
  })

  it('下拉口径常量非空', () => {
    expect(statusOptions.length).toBeGreaterThan(0)
    expect(envTypeOptions).toContain('生产')
    expect(assetTypeOptions.length).toBeGreaterThan(0)
  })
})

describe('filterCloudRegions（T-086 资源统计口径）', () => {
  it('仅保留区域名含「云」的区域', () => {
    expect(filterCloudRegions(['南宁云', '华东云', '华南', '北京'])).toEqual(['南宁云', '华东云'])
  })
  it('无匹配返回空数组', () => {
    expect(filterCloudRegions(['华南', '北京'])).toEqual([])
  })
  it('空输入安全', () => {
    expect(filterCloudRegions()).toEqual([])
    expect(filterCloudRegions([])).toEqual([])
    expect(filterCloudRegions([null, '', '南宁云'])).toEqual(['南宁云'])
  })
})
