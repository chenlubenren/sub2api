<template>
  <!-- 后台内嵌形态:?embedded=1 且已登录,套完整后台布局 -->
  <AppLayout v-if="isEmbedded">
    <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" embedded standard-only />
  </AppLayout>

  <!-- 独立形态:自带导航条(logo/站名 + 登录/回后台) -->
  <div v-else class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <PlazaNavBar />
    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" standard-only />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import PlazaNavBar from '@/components/modelPlaza/PlazaNavBar.vue'
import ModelPlazaContent from '@/components/modelPlaza/ModelPlazaContent.vue'
import { getModelPlaza, getModelPricing, type ModelPlazaResponse } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { ModelPlazaGroup, PlazaModel } from '@/api/modelPlaza'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

// embedded=1 但未登录(如转发的链接)自动降级为独立形态。
const isEmbedded = computed(() => (route.query.embedded === '1' || route.path === '/model-pricing') && authStore.isAuthenticated)

const data = ref<ModelPlazaResponse | null>(null)
const loading = ref(true)
const loadFailed = ref(false)

// 用户价目只展示 ChatGPT 与 Claude Code 的常用模型，并统一按标准 1x
// 展示。管理员仍可在后台为渠道/分组维护真实计费价格。
const STANDARD_MODEL_ORDER = [
  'gpt-6-astra',
  'gpt-6.1-sol',
  'gpt-6-sol',
  'gpt-5.6-sol',
  'gpt-5.5',
  'claude-opus-4-8',
  'claude-opus-4-6',
  'claude-sonnet-4-6',
  'claude-haiku-4-5'
]
const STANDARD_MODEL_INDEX = new Map(STANDARD_MODEL_ORDER.map((name, index) => [name, index]))

function normalizeStandardPricing(response: ModelPlazaResponse): ModelPlazaResponse {
  const byPlatform = new Map<string, Map<string, PlazaModel>>()
  for (const group of response.groups) {
    for (const model of group.models) {
      if (!STANDARD_MODEL_INDEX.has(model.name)) continue
      const platform = model.platform === 'anthropic' ? 'anthropic' : 'openai'
      const models = byPlatform.get(platform) || new Map<string, PlazaModel>()
      const standardModel = toStandardModel(model)
      const previous = models.get(model.name)
      // 同一模型可能来自多个分组；标准价目只保留一行，优先保留有价格的条目。
      if (!previous || (!previous.pricing && standardModel.pricing)) models.set(model.name, standardModel)
      byPlatform.set(platform, models)
    }
  }

  const makeGroup = (platform: 'openai' | 'anthropic', name: string, id: number): ModelPlazaGroup | null => {
    const models = [...(byPlatform.get(platform)?.values() || [])].sort(
      (left, right) => (STANDARD_MODEL_INDEX.get(left.name) ?? Number.MAX_SAFE_INTEGER) - (STANDARD_MODEL_INDEX.get(right.name) ?? Number.MAX_SAFE_INTEGER)
    )
    if (!models.length) return null
    return {
      id,
      name,
      description: '标准 1x 价格，按每百万 token 计价。',
      platform,
      subscription_type: 'standard',
      rate_multiplier: 1,
      peak_rate_enabled: false,
      peak_start: '',
      peak_end: '',
      peak_rate_multiplier: 1,
      is_exclusive: false,
      image_rate_independent: false,
      image_rate_multiplier: 1,
      long_context_pricing_enabled: true,
      models
    }
  }

  return {
    description: 'ChatGPT 与 Claude Code 常用模型的标准 1x 价格。',
    groups: [
      makeGroup('openai', 'OpenAI GPT 常用模型', 0),
      makeGroup('anthropic', 'Anthropic Claude Code 常用模型', 1)
    ].filter((group): group is ModelPlazaGroup => group !== null)
  }
}

function toStandardModel(model: PlazaModel): PlazaModel {
  const official = model.official_pricing
  if (!official) return model
  return {
    ...model,
    // 官方参考价本身就是无渠道、无分组倍率的标准 1x 价格。
    pricing: {
      billing_mode: 'token',
      input_price: official.input_price,
      output_price: official.output_price,
      cache_write_price: official.cache_write_price,
      cache_write_1h_price: official.cache_write_1h_price ?? null,
      cache_read_price: official.cache_read_price,
      max_reasoning_effort_multiplier: model.pricing?.max_reasoning_effort_multiplier ?? null,
      image_input_price: null,
      image_output_price: null,
      per_request_price: null,
      intervals: official.intervals || []
    }
  }
}

onMounted(async () => {
  // 独立形态导航条需要站点名/Logo;有 __APP_CONFIG__ 注入时同步命中缓存。
  void appStore.fetchPublicSettings()
  try {
    const response = await (route.path === '/model-pricing' ? getModelPricing() : getModelPlaza())
    data.value = route.path === '/model-pricing' ? normalizeStandardPricing(response) : response
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
})
</script>
