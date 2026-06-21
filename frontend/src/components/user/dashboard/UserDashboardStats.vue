<template>
  <div class="space-y-4">
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div v-if="!isSimple" class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
            <svg class="h-5 w-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
            </svg>
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.balance') }}</p>
            <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">${{ formatBalance(balance) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('common.available') }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
            <Icon name="key" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.apiKeys') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stats?.total_api_keys || 0 }}</p>
            <p class="text-xs text-green-600 dark:text-green-400">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
            <Icon name="chart" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.todayRequests') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stats?.today_requests || 0 }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
            <Icon name="dollar" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.todayCost') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              <span class="text-purple-600 dark:text-purple-400">${{ formatCost(stats?.today_actual_cost || 0) }}</span>
              <span class="text-sm font-normal text-gray-500 dark:text-gray-400"> / 成本：${{ formatCost(todayAccountCost) }}</span>
            </p>
            <p class="text-xs">
              <span class="text-gray-500 dark:text-gray-400">{{ t('common.total') }}: </span>
              <span class="text-purple-600 dark:text-purple-400">${{ formatCost(stats?.total_actual_cost || 0) }}</span>
              <span class="text-gray-500 dark:text-gray-400"> / 成本：${{ formatCost(totalAccountCost) }}</span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
            <Icon name="cube" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.todayTokens') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats?.today_tokens || 0) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
            <Icon name="database" size="md" class="text-indigo-600 dark:text-indigo-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.totalTokens') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats?.total_tokens || 0) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-violet-100 p-2 dark:bg-violet-900/30">
            <Icon name="bolt" size="md" class="text-violet-600 dark:text-violet-400" :stroke-width="2" />
          </div>
          <div class="flex-1">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.performance') }}</p>
            <div class="flex items-baseline gap-2">
              <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats?.rpm || 0) }}</p>
              <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">{{ formatTokens(stats?.tpm || 0) }}</p>
              <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
            <Icon name="clock" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.avgResponse') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.averageTime') }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="mx-auto w-full max-w-[1380px]">
      <div class="card px-6 py-5">
        <p class="flex flex-wrap items-center gap-x-3 gap-y-2 text-base font-semibold text-gray-900 dark:text-white">
          <span>算力站本日已为您节省</span>
          <span class="text-3xl font-black text-emerald-600 dark:text-emerald-400">￥{{ formatRmb(todaySavedRmb) }}</span>
          <span>，已累计为您节省</span>
          <span class="text-3xl font-black text-violet-600 dark:text-violet-400">￥{{ formatRmb(totalSavedRmb) }}</span>
          <span>。</span>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()

const { t } = useI18n()

function getNumericStat(key: string): number {
  const value = (props.stats as unknown as Record<string, unknown> | undefined)?.[key]
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}

const todayAccountCost = computed(() => getNumericStat('today_account_cost'))
const totalAccountCost = computed(() => getNumericStat('total_account_cost'))
const todaySavedRmb = computed(() => Math.max(0, todayAccountCost.value * 6.7 - (props.stats?.today_actual_cost || 0)))
const totalSavedRmb = computed(() => Math.max(0, totalAccountCost.value * 6.7 - (props.stats?.total_actual_cost || 0)))

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatRmb = (c: number) => c.toFixed(2)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => (ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`)
</script>
