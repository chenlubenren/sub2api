<template>
  <div class="space-y-6">
    <div class="card p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('dashboard.timeRange') }}:</span>
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:startDate="$emit('update:startDate', $event)"
            @update:endDate="$emit('update:endDate', $event)"
            @change="$emit('dateRangeChange', $event)"
          />
        </div>
        <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary">
          {{ t('common.refresh') }}
        </button>
        <div class="ml-auto flex items-center gap-2">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('dashboard.granularity') }}:</span>
          <div class="w-28">
            <Select
              :model-value="granularity"
              :options="[{value:'day', label:t('dashboard.day')}, {value:'hour', label:t('dashboard.hour')}]"
              @update:model-value="$emit('update:granularity', $event)"
              @change="$emit('granularityChange')"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <div class="card relative overflow-hidden border-2 border-[#101010] bg-white p-4 shadow-[4px_4px_0_#101010]">
        <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center bg-white/70 backdrop-blur-sm">
          <LoadingSpinner size="md" />
        </div>
        <h3 class="mb-4 text-sm font-semibold text-[#101010]">{{ t('dashboard.modelDistribution') }}</h3>

        <div v-if="models.length" class="space-y-4">
          <div class="space-y-2 border-2 border-[#101010] bg-[#f8f8f8] p-3">
            <div
              v-for="model in pixelModelRows"
              :key="model.model"
              class="grid grid-cols-[minmax(0,120px)_1fr_auto] items-center gap-3 text-xs"
            >
              <span class="truncate font-semibold text-[#101010]" :title="model.model">{{ model.model }}</span>
              <div class="h-5 border-2 border-[#101010] bg-white">
                <div
                  class="h-full border-r-2 border-[#101010]"
                  :style="{ width: `${model.percent}%`, backgroundColor: model.color }"
                ></div>
              </div>
              <span class="font-mono text-[#101010]/65">{{ formatTokens(model.total_tokens) }}</span>
            </div>
          </div>

          <div class="max-h-48 overflow-y-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-gray-500 dark:text-gray-400">
                  <th class="pb-2 text-left">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.actual') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="border-t border-gray-100 dark:border-gray-700">
                  <td class="max-w-[100px] truncate py-1.5 font-medium text-gray-900 dark:text-white" :title="model.model">{{ model.model }}</td>
                  <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">{{ formatNumber(model.requests) }}</td>
                  <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="py-1.5 text-right text-green-600 dark:text-green-400">￥{{ formatCost(model.actual_cost) }}</td>
                  <td class="py-1.5 text-right text-gray-400 dark:text-gray-500">￥{{ formatCost(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else class="flex h-48 items-center justify-center text-sm text-gray-500">
          {{ t('dashboard.noDataAvailable') }}
        </div>
      </div>

      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'

const props = defineProps<{
  loading: boolean
  startDate: string
  endDate: string
  granularity: string
  trend: TrendDataPoint[]
  models: ModelStat[]
}>()

defineEmits([
  'update:startDate',
  'update:endDate',
  'update:granularity',
  'dateRangeChange',
  'granularityChange',
  'refresh'
])

const { t } = useI18n()

const pixelColors = ['#3A5BA0', '#2D7D46', '#D4A533', '#A83232', '#6B6B6B']

const pixelModelRows = computed(() => {
  const rows = [...(props.models || [])].sort((a, b) => b.total_tokens - a.total_tokens).slice(0, 8)
  const maxTokens = Math.max(...rows.map((model) => model.total_tokens), 1)
  return rows.map((model, index) => ({
    ...model,
    color: pixelColors[index % pixelColors.length],
    percent: Math.max(4, Math.round((model.total_tokens / maxTokens) * 100))
  }))
})
</script>
