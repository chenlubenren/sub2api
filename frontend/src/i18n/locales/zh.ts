import legacy from './zh-legacy'
import modular from './zh/index'

type LocaleMessages = Record<string, any>

function isRecord(value: unknown): value is LocaleMessages {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}

function mergeLocaleMessages(primary: LocaleMessages, supplemental: LocaleMessages): LocaleMessages {
  if (!isRecord(primary) || !isRecord(supplemental)) {
    return primary
  }

  const merged: LocaleMessages = { ...primary }
  for (const [key, value] of Object.entries(supplemental)) {
    if (!(key in merged)) {
      merged[key] = value
    } else if (isRecord(merged[key]) && isRecord(value)) {
      merged[key] = mergeLocaleMessages(merged[key], value)
    }
  }
  return merged
}

export default mergeLocaleMessages(legacy, modular)
