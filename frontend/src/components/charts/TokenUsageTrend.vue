<template>
  <div class="card border-2 border-[#101010] bg-white p-4 shadow-[4px_4px_0_#101010]">
    <h3 class="mb-4 text-sm font-semibold text-[#101010]">
      {{ t('admin.dashboard.tokenUsageTrend') }}
    </h3>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0" class="space-y-3">
      <div class="relative h-48 border-2 border-[#101010] bg-[#f8f8f8] p-3">
        <div class="absolute inset-3 bg-[linear-gradient(rgba(16,16,16,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(16,16,16,0.08)_1px,transparent_1px)] bg-[size:18px_18px]"></div>
        <svg class="relative h-full w-full overflow-visible" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
          <polyline
            v-for="series in pixelSeries"
            :key="series.key"
            :points="series.points"
            fill="none"
            :stroke="series.color"
            stroke-width="3"
            stroke-linejoin="miter"
            stroke-linecap="square"
          />
          <g v-for="series in pixelSeries" :key="`${series.key}-points`">
            <rect
              v-for="point in series.pointRects"
              :key="`${series.key}-${point.x}-${point.y}`"
              :x="point.x - 1.8"
              :y="point.y - 1.8"
              width="3.6"
              height="3.6"
              :fill="series.color"
              stroke="#101010"
              stroke-width="0.7"
            />
          </g>
        </svg>
      </div>

      <div class="grid gap-2 sm:grid-cols-2">
        <div
          v-for="series in pixelSeries"
          :key="`${series.key}-legend`"
          class="flex items-center justify-between gap-2 border-2 border-[#101010] bg-[#f8f8f8] px-3 py-2 text-xs"
        >
          <span class="flex items-center gap-2 font-semibold text-[#101010]">
            <span class="h-3 w-3 border border-[#101010]" :style="{ backgroundColor: series.color }"></span>
            {{ series.label }}
          </span>
          <span class="font-mono text-[#101010]/65">{{ formatTokens(series.total) }}</span>
        </div>
      </div>
    </div>
    <div
      v-else
      class="flex h-48 items-center justify-center text-sm text-gray-500"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

const { t } = useI18n()

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
}>()

const chartRows = computed(() => props.trendData.slice(-18))

const maxTokens = computed(() => {
  const values = chartRows.value.flatMap((d) => [
    d.input_tokens,
    d.output_tokens,
    d.cache_creation_tokens,
    d.cache_read_tokens
  ])
  return Math.max(...values, 1)
})

const seriesConfig = [
  { key: 'input', label: 'Input', color: '#3A5BA0', getValue: (d: TrendDataPoint) => d.input_tokens },
  { key: 'output', label: 'Output', color: '#2D7D46', getValue: (d: TrendDataPoint) => d.output_tokens },
  { key: 'cache', label: 'Cache', color: '#D4A533', getValue: (d: TrendDataPoint) => d.cache_creation_tokens + d.cache_read_tokens }
]

const pixelSeries = computed(() => {
  const rowCount = Math.max(chartRows.value.length - 1, 1)
  return seriesConfig.map((series) => {
    const pointRects = chartRows.value.map((row, index) => {
      const x = chartRows.value.length === 1 ? 50 : (index / rowCount) * 100
      const y = 96 - (series.getValue(row) / maxTokens.value) * 86
      return { x, y }
    })
    return {
      ...series,
      pointRects,
      points: pointRects.map((point) => `${point.x.toFixed(2)},${point.y.toFixed(2)}`).join(' '),
      total: chartRows.value.reduce((sum, row) => sum + series.getValue(row), 0)
    }
  })
})

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}
</script>
