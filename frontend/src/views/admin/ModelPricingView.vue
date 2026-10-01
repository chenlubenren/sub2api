<template>
  <AppLayout>
    <div class="mx-auto max-w-[1380px] space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div class="mb-2 flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-primary-600">
            <Icon name="chart" size="sm" />计费配置
          </div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">模型定价</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            全局标准 1x 价格。保存后，所有对应模型的后续计费立即按此价格执行。
          </p>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <span class="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300">
            <span class="h-2 w-2 rounded-full" :class="savedAt ? 'bg-emerald-500' : 'bg-amber-500'" />
            {{ savedAt ? `已保存 ${savedAt}` : '尚未保存本页修改' }}
          </span>
          <button class="btn btn-primary" :disabled="saving || loading || !pricing.length" @click="save">
            <Icon name="check" size="sm" />{{ saving ? '保存中' : '保存定价' }}
          </button>
        </div>
      </header>

      <div v-if="loading" class="flex h-72 items-center justify-center"><LoadingSpinner /></div>
      <div v-else-if="!pricing.length" class="card p-8 text-sm text-gray-500">暂无可配置的标准模型。</div>
      <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="item in pricing"
          :key="item.model"
          class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="mb-4 flex items-start justify-between gap-2">
            <div>
              <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.label }}</h3>
              <p class="mt-1 font-mono text-xs text-gray-500">{{ item.model }}</p>
            </div>
            <span
              class="rounded-full px-2 py-1 text-[11px] font-medium"
              :class="item.platform === 'openai' ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-300' : 'bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-300'"
            >{{ item.platform === 'openai' ? 'OpenAI' : 'Anthropic' }}</span>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <label v-for="field in fields" :key="field.key" class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ field.label }} ($ / 1M)
              <input
                v-model.number="item[field.key]"
                type="number"
                min="0"
                step="0.01"
                class="input mt-1 w-full text-sm text-gray-900 dark:text-white"
                @input="savedAt = ''"
              />
            </label>
          </div>
        </article>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import modelPricingAPI, {
  type GlobalModelPricing,
  type ModelPricingPlatform
} from '@/api/admin/modelPricing'
import { useAppStore } from '@/stores/app'

type PriceKey = 'input' | 'output' | 'cacheWrite' | 'cacheRead'
type Card = {
  model: string
  label: string
  platform: ModelPricingPlatform
  input: number
  output: number
  cacheWrite: number
  cacheRead: number
}

const fields: Array<{ key: PriceKey; label: string }> = [
  { key: 'input', label: '输入' },
  { key: 'output', label: '输出' },
  { key: 'cacheWrite', label: '缓存写入' },
  { key: 'cacheRead', label: '缓存读取' }
]

const appStore = useAppStore()
const pricing = ref<Card[]>([])
const loading = ref(true)
const saving = ref(false)
const savedAt = ref('')

const toMillion = (value: number) => Number((value * 1_000_000).toFixed(4))
const toToken = (value: number) => value / 1_000_000

function toCard(item: GlobalModelPricing): Card {
  return {
    model: item.model,
    label: item.label,
    platform: item.platform,
    input: toMillion(item.input_price),
    output: toMillion(item.output_price),
    cacheWrite: toMillion(item.cache_write_price),
    cacheRead: toMillion(item.cache_read_price)
  }
}

function toRequest(item: Card): GlobalModelPricing {
  return {
    model: item.model,
    label: item.label,
    platform: item.platform,
    input_price: toToken(item.input),
    output_price: toToken(item.output),
    cache_write_price: toToken(item.cacheWrite),
    cache_read_price: toToken(item.cacheRead)
  }
}

async function load() {
  loading.value = true
  try {
    pricing.value = (await modelPricingAPI.getGlobalModelPricing()).map(toCard)
  } catch (error) {
    console.error(error)
    appStore.showError('无法加载模型定价')
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    pricing.value = (await modelPricingAPI.updateGlobalModelPricing(pricing.value.map(toRequest))).map(toCard)
    savedAt.value = new Date().toLocaleTimeString()
    appStore.showSuccess('全局模型定价已保存，并已用于后续计费')
  } catch (error) {
    console.error(error)
    appStore.showError('保存模型定价失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
