<template>
  <AppLayout>
    <div class="mx-auto max-w-[1380px] space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-4">
        <div><div class="mb-2 flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-primary-600"><Icon name="chart" size="sm" />计费配置</div><h1 class="text-2xl font-semibold text-gray-900 dark:text-white">模型定价</h1><p class="mt-1 text-sm text-gray-500 dark:text-gray-400">按供应商查看输入、输出、缓存写入和缓存读取价格。保存后用于该渠道后续计费。</p></div>
        <div class="flex flex-wrap items-center justify-end gap-2"><span class="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300"><span class="h-2 w-2 rounded-full" :class="savedAt ? 'bg-emerald-500' : 'bg-amber-500'" />{{ savedAt ? `已保存 ${savedAt}` : '尚未保存本页修改' }}</span><select v-model.number="channelId" class="input w-56"><option v-for="item in channels" :key="item.id" :value="item.id">{{ item.name }}</option></select><button class="btn btn-primary" :disabled="saving || !channel" @click="save"><Icon name="check" size="sm" />{{ saving ? '保存中' : '保存定价' }}</button></div>
      </header>
      <div v-if="loading" class="flex h-72 items-center justify-center"><LoadingSpinner /></div>
      <div v-else-if="!channel" class="card p-8 text-sm text-gray-500">请先创建一个渠道，再为渠道配置模型定价。</div>
      <div v-else class="space-y-5"><section v-for="section in sections" :key="section.platform" class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800"><div class="flex items-center justify-between border-b border-gray-100 bg-gray-50/80 px-5 py-4 dark:border-dark-700 dark:bg-dark-900/40"><div><h2 class="font-semibold text-gray-900 dark:text-white">{{ section.label }}</h2><p class="mt-1 text-xs text-gray-500">{{ section.items.length }} 个模型 · $ / 1M tokens</p></div><span class="rounded-full px-3 py-1 text-xs font-medium" :class="section.platform === 'openai' ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-300' : 'bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-300'">{{ section.platform === 'openai' ? 'OpenAI' : 'Anthropic' }}</span></div><div class="grid grid-cols-1 gap-4 p-5 md:grid-cols-2 xl:grid-cols-3"><article v-for="item in section.items" :key="item.model" class="rounded-lg border border-gray-200 p-4 transition-shadow hover:shadow-md dark:border-dark-600"><div class="mb-4 flex items-start justify-between gap-2"><div><h3 class="font-semibold text-gray-900 dark:text-white">{{ item.label }}</h3><p class="mt-1 font-mono text-xs text-gray-500">{{ item.model }}</p></div><span class="text-xs text-gray-400">{{ item.platform === 'openai' ? 'GPT' : 'Claude' }}</span></div><div class="grid grid-cols-2 gap-3"><label v-for="field in fields" :key="field.key" class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ field.label }}<input v-model.number="item[field.key]" type="number" min="0" step="0.01" class="input mt-1 w-full text-sm text-gray-900 dark:text-white" /></label></div></article></div></section></div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import channelsAPI, { type Channel, type ChannelModelPricing } from '@/api/admin/channels'
import { useAppStore } from '@/stores/app'

type PriceKey = 'input' | 'output' | 'cacheWrite' | 'cacheRead'
type Platform = 'openai' | 'anthropic'
type Card = { model: string; label: string; platform: Platform; input: number; output: number; cacheWrite: number; cacheRead: number }
const models: Array<[string, string, Platform]> = [
  ['gpt-5.5', 'GPT-5.5', 'openai'],
  ['gpt-5.6-sol', 'GPT-5.6 Sol', 'openai'],
  ['gpt-5.6-terra', 'GPT-5.6 Terra', 'openai'],
  ['gpt-5.6-luna', 'GPT-5.6 Luna', 'openai'],
  ['gpt-6-astra', 'GPT-6 Astra', 'openai'],
  ['gpt-6-sol', 'GPT-6 Sol', 'openai'],
  ['gpt-6-luna', 'GPT-6 Luna', 'openai'],
  ['gpt-6.1-sol', 'GPT-6.1 Sol', 'openai'],
  ['claude-opus-5-5', 'Claude Opus 5.5', 'anthropic'],
  ['claude-opus-5', 'Claude Opus 5', 'anthropic'],
  ['claude-sonnet-5-5', 'Claude Sonnet 5.5', 'anthropic'],
  ['claude-sonnet-5', 'Claude Sonnet 5', 'anthropic'],
  ['claude-opus-4-8', 'Claude Opus 4.8', 'anthropic'],
  ['claude-opus-4-7', 'Claude Opus 4.7', 'anthropic'],
  ['claude-opus-4-6', 'Claude Opus 4.6', 'anthropic'],
  ['claude-sonnet-4-6', 'Claude Sonnet 4.6', 'anthropic'],
  ['claude-haiku-4-5', 'Claude Haiku 4.5', 'anthropic']
]
const fields: Array<{ key: PriceKey; label: string }> = [{ key: 'input', label: '输入' }, { key: 'output', label: '输出' }, { key: 'cacheWrite', label: '缓存写入' }, { key: 'cacheRead', label: '缓存读取' }]
const appStore = useAppStore(); const channels = ref<Channel[]>([]); const channelId = ref<number>(); const pricing = ref<Card[]>([]); const loading = ref(true); const saving = ref(false); const savedAt = ref('')
const channel = computed(() => channels.value.find(item => item.id === channelId.value))
const sections = computed(() => (['openai', 'anthropic'] as Platform[]).map(platform => ({ platform, label: platform === 'openai' ? 'OpenAI GPT 系列' : 'Anthropic Claude 系列', items: pricing.value.filter(item => item.platform === platform) })).filter(section => section.items.length))
const toMillion = (value: number | null | undefined) => Number(((value || 0) * 1_000_000).toFixed(4)); const toToken = (value: number) => value / 1_000_000
async function buildCards() { const selected = channel.value; if (!selected) { pricing.value = []; return }; pricing.value = await Promise.all(models.map(async ([model, label, platform]) => { const saved = selected.model_pricing.find(rule => rule.models.includes(model)); if (saved) return { model, label, platform, input: toMillion(saved.input_price), output: toMillion(saved.output_price), cacheWrite: toMillion(saved.cache_write_price), cacheRead: toMillion(saved.cache_read_price) }; const base = await channelsAPI.getModelDefaultPricing(model); return { model, label, platform, input: toMillion(base.input_price), output: toMillion(base.output_price), cacheWrite: toMillion(base.cache_write_price), cacheRead: toMillion(base.cache_read_price) } })) }
async function load() { loading.value = true; try { const response = await channelsAPI.list(1, 100); channels.value = response.items; channelId.value = channels.value[0]?.id; await buildCards() } catch (error) { console.error(error); appStore.showError('无法加载渠道模型定价') } finally { loading.value = false } }
watch(channelId, () => { savedAt.value = ''; void buildCards() })
async function save() { if (!channel.value) return; saving.value = true; try { const retained = channel.value.model_pricing.filter(rule => !models.some(([model]) => rule.models.includes(model))); const edited: ChannelModelPricing[] = pricing.value.map(item => ({ platform: item.platform, models: [item.model], billing_mode: 'token', input_price: toToken(item.input), output_price: toToken(item.output), cache_write_price: toToken(item.cacheWrite), cache_read_price: toToken(item.cacheRead), image_input_price: null, image_output_price: null, per_request_price: null, intervals: [], time_pricing: null })); const updated = await channelsAPI.update(channel.value.id, { model_pricing: [...retained, ...edited] }); channels.value = channels.value.map(item => item.id === updated.id ? updated : item); savedAt.value = new Date().toLocaleTimeString(); appStore.showSuccess(`已保存 ${updated.name} 的模型定价，并用于后续计费`) } catch (error) { console.error(error); appStore.showError('保存模型定价失败') } finally { saving.value = false } }
onMounted(load)
</script>
