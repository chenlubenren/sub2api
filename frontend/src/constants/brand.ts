export const DEFAULT_SITE_NAME = '助研算力供应中心'
export const LEGACY_SITE_NAME = 'Sub2API'
export const PREVIOUS_SITE_NAME = '助研中转站'
export const DEFAULT_SITE_SUBTITLE = '拒绝掺水，不转卖用户隐私，高度安全，超低价玩转codex'
export const DEFAULT_DOCUMENT_TITLE_SUFFIX = 'AI API Gateway'

const LEGACY_SUBTITLES = new Set([
  'AI API Gateway Platform',
  'Subscription to API',
  'Subscription to API Conversion Platform',
  '拒绝掺水，不转卖用户隐私，高度安全，1折玩转 Codex',
  '拒绝掺水，不转卖用户隐私，高度安全，超低价玩转codex'
])

export function normalizeSiteName(value?: string | null): string {
  const name = value?.trim()
  if (!name || name === LEGACY_SITE_NAME || name === PREVIOUS_SITE_NAME) return DEFAULT_SITE_NAME
  return name
}

export function normalizeSiteSubtitle(value?: string | null): string {
  const subtitle = value?.trim()
  if (!subtitle || LEGACY_SUBTITLES.has(subtitle)) return DEFAULT_SITE_SUBTITLE
  return subtitle
}
