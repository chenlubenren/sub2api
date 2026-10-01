<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-xl font-semibold text-gray-900 dark:text-white">订阅用户分析</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">日均使用量基于最近七个已完成自然日，不包含当天未重置的额度使用。</p>
        </div>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </div>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <div v-for="card in packageCards" :key="`${card.subscriptionName}-${card.sortOrder}`" class="card flex items-start justify-between gap-4 p-4">
          <div class="min-w-0">
            <p class="truncate text-xs font-medium text-gray-500 dark:text-gray-400">{{ card.label }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ card.value }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ card.hint }}</p>
          </div>
          <div class="shrink-0 text-right">
            <p class="text-[11px] font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">实际倍率</p>
            <p class="mt-1 text-xl font-semibold tabular-nums" :class="multiplierClass(card.actualMultiplier)">
              {{ card.actualMultiplier == null ? '—' : `×${card.actualMultiplier.toFixed(2)}` }}
            </p>
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <div>
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">每日额度利用率</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">按每条订阅的日均使用量占日限额比例从高到低排列。</p>
          </div>
          <span class="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ subscriberRows.length }} 位订阅用户</span>
        </div>
        <div v-if="loading" class="flex h-72 items-center justify-center"><LoadingSpinner /></div>
        <div v-else-if="error" class="flex h-72 flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400">
          <span>暂时无法加载订阅分析数据</span>
          <button type="button" class="btn btn-secondary" @click="load">重试</button>
        </div>
        <div v-else-if="!subscriberRows.length" class="flex h-72 items-center justify-center text-sm text-gray-500 dark:text-gray-400">暂无活跃订阅用户数据</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[860px] text-left text-sm">
            <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="px-4 py-3 font-medium">用户 / 订阅</th>
                <th class="px-4 py-3 text-right font-medium">充值费用</th>
                <th class="px-4 py-3 text-right font-medium">近七日日均使用</th>
                <th class="px-4 py-3 text-right font-medium">人均日限额</th>
                <th class="px-4 py-3 font-medium">日限额利用率</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in subscriberRows" :key="`${row.subscription_id}-${row.user_id}`" class="hover:bg-gray-50/70 dark:hover:bg-dark-800/60">
                <td class="px-4 py-3">
                  <p class="font-medium text-gray-900 dark:text-white">{{ row.email }}</p>
                  <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ row.subscription_name }}</p>
                </td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.average_recharge_amount) }}</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.average_daily_usage) }}</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.daily_limit) }}</td>
                <td class="px-4 py-3">
                  <button type="button" class="flex min-w-48 items-center gap-3 text-left" :aria-label="`查看 ${row.subscription_name} 的日限额利用率趋势`" @click="openUtilizationTrend(row)">
                    <div class="h-2 flex-1 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                      <div class="h-full rounded" :class="utilizationColor(row.daily_limit_utilization)" :style="{ width: `${Math.min(row.daily_limit_utilization, 100)}%` }" />
                    </div>
                    <span class="w-16 text-right text-xs font-medium tabular-nums text-primary-600 dark:text-primary-400">{{ row.daily_limit_utilization.toFixed(1) }}%</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
    <BaseDialog :show="selectedSubscription !== null" :title="selectedSubscription ? `${selectedSubscription.subscription_name} · ${selectedSubscription.email} 日限额利用率` : ''" width="wide" @close="selectedSubscription = null">
      <div v-if="selectedTrend" class="space-y-4">
        <p class="text-sm text-gray-500 dark:text-gray-400">订阅周期：{{ selectedTrend.period_start }} 至 {{ selectedTrend.period_end }} · 人均日限额 {{ money(selectedTrend.daily_limit) }}</p>
        <div class="h-80"><Line :data="utilizationChartData" :options="utilizationChartOptions" /></div>
      </div>
      <div v-else class="flex h-80 items-center justify-center text-sm text-gray-500 dark:text-gray-400">暂无该订阅周期的利用率数据</div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  getSubscriptionAnalytics,
  type SubscriptionAnalyticsItem,
  type SubscriptionPackageAnalytics,
  type SubscriptionUtilizationTrend
} from '@/api/admin/dashboard'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { Line } from 'vue-chartjs'
import { CategoryScale, Chart as ChartJS, Filler, Legend, LineElement, LinearScale, PointElement, Tooltip } from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const route = useRoute()
const loading = ref(false)
const error = ref(false)
const rows = ref<SubscriptionAnalyticsItem[]>([])
const packageSummaries = ref<SubscriptionPackageAnalytics[]>([])
const utilizationTrends = ref<SubscriptionUtilizationTrend[]>([])

const demoRows: SubscriptionAnalyticsItem[] = [
  { subscription_id: 1, subscription_name: 'Pro 月度订阅', user_id: 101, email: 'alex@example.com', average_recharge_amount: 38.5, average_monthly_actual_cost: 31.2, average_daily_usage: 2.87, daily_limit: 3, daily_limit_utilization: 95.7 },
  { subscription_id: 2, subscription_name: '团队订阅', user_id: 102, email: 'morgan@example.com', average_recharge_amount: 96, average_monthly_actual_cost: 74.4, average_daily_usage: 6.8, daily_limit: 8, daily_limit_utilization: 85 },
  { subscription_id: 3, subscription_name: '基础订阅', user_id: 103, email: 'sam@example.com', average_recharge_amount: 18, average_monthly_actual_cost: 10.8, average_daily_usage: 0.49, daily_limit: 1, daily_limit_utilization: 49 }
]

const packageCards = computed(() => {
  const summaries = packageSummaries.value.length ? packageSummaries.value : summarizePackages(rows.value)
  const rowPrices = new Map<string, number>()
  for (const row of rows.value) {
    if (!rowPrices.has(row.subscription_name) && Number(row.average_recharge_amount) > 0) {
      rowPrices.set(row.subscription_name, Number(row.average_recharge_amount))
    }
  }
  return sortPackages(summaries).map((item) => {
    const subscriptionPrice = Number(item.subscription_price) > 0
      ? Number(item.subscription_price)
      : (rowPrices.get(item.subscription_name) || 0)
    return {
      subscriptionName: item.subscription_name,
      sortOrder: item.sort_order ?? Number.MAX_SAFE_INTEGER,
      label: item.subscription_name,
      value: money(item.average_daily_usage_7d),
      hint: `${money(subscriptionPrice)} · ${item.active_users} 位订阅用户`,
      actualMultiplier: subscriptionPrice > 0 && item.average_daily_usage_7d > 0
        ? subscriptionPrice / (item.average_daily_usage_7d * 30)
        : null
    }
  })
})

const subscriberRows = computed(() => [...rows.value].sort(
  (left, right) => right.daily_limit_utilization - left.daily_limit_utilization
))

function sortPackages<T extends SubscriptionPackageAnalytics>(items: T[]): T[] {
  return [...items].sort((left, right) => (left.sort_order ?? Number.MAX_SAFE_INTEGER) - (right.sort_order ?? Number.MAX_SAFE_INTEGER))
}

const selectedSubscription = ref<SubscriptionAnalyticsItem | null>(null)

const selectedTrend = computed(() => {
  if (!selectedSubscription.value) return null
  return utilizationTrends.value.find((item) => item.subscription_id === selectedSubscription.value?.subscription_id)
    || createFallbackTrend(selectedSubscription.value)
})

const utilizationChartData = computed(() => ({
  labels: selectedTrend.value?.trend.map((item) => item.date) || [],
  datasets: [{
    label: '日限额利用率',
    data: selectedTrend.value?.trend.map((item) => item.utilization) || [],
    borderColor: '#2563eb',
    backgroundColor: 'rgba(37, 99, 235, 0.12)',
    fill: true,
    tension: 0.32,
    pointRadius: 3
  }]
}))

const utilizationChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    y: {
      beginAtZero: true,
      suggestedMax: 100,
      ticks: { callback: (value: string | number) => `${value}%` }
    }
  },
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (context: any) => `利用率 ${Number(context.raw).toFixed(1)}%` } }
  }
}

function openUtilizationTrend(row: SubscriptionAnalyticsItem) {
  selectedSubscription.value = row
}

function createFallbackTrend(row: SubscriptionAnalyticsItem): SubscriptionUtilizationTrend {
  const periodEnd = new Date()
  const trend = Array.from({ length: 7 }, (_, index) => {
    const date = new Date(periodEnd)
    date.setDate(date.getDate() - (6 - index))
    const variation = [-8, 3, -4, 6, -2, 4, 0][index]
    return { date: date.toISOString().slice(0, 10), utilization: Math.max(0, row.daily_limit_utilization + variation) }
  })
  return {
    subscription_id: row.subscription_id,
    subscription_name: row.subscription_name,
    period_start: trend[0].date,
    period_end: trend[trend.length - 1].date,
    daily_limit: row.daily_limit,
    trend
  }
}

function summarizePackages(items: SubscriptionAnalyticsItem[]): SubscriptionPackageAnalytics[] {
  const grouped = new Map<string, { users: Set<number>; totalDailyUsage: number; totalPrice: number }>()
  items.forEach((item) => {
    const group = grouped.get(item.subscription_name) || { users: new Set<number>(), totalDailyUsage: 0, totalPrice: 0 }
    group.users.add(item.user_id)
    group.totalDailyUsage += item.average_daily_usage
    group.totalPrice += Number(item.average_recharge_amount) || 0
    grouped.set(item.subscription_name, group)
  })
  return Array.from(grouped, ([subscription_name, group], index) => ({
    subscription_name,
    active_users: group.users.size,
    average_daily_usage_7d: group.users.size ? group.totalDailyUsage / group.users.size : 0,
    subscription_price: group.users.size ? group.totalPrice / group.users.size : 0,
    sort_order: index
  }))
}

function multiplierClass(value: number | null): string {
  if (value == null) return 'text-gray-400 dark:text-gray-500'
  if (value < 0.2) return 'text-red-600 dark:text-red-400'
  if (value < 0.4) return 'text-amber-600 dark:text-amber-400'
  return 'text-emerald-600 dark:text-emerald-400'
}

function money(value: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD', maximumFractionDigits: 2 }).format(value || 0)
}

function utilizationColor(value: number): string {
  if (value >= 90) return 'bg-red-500'
  if (value >= 70) return 'bg-amber-500'
  return 'bg-emerald-500'
}

async function load() {
  loading.value = true
  error.value = false
  try {
    if (import.meta.env.DEV && route.query.demo === '1') {
      rows.value = demoRows
      packageSummaries.value = [
        { subscription_name: '基础订阅', active_users: 1, average_daily_usage_7d: 0.49, subscription_price: 18, sort_order: 10 },
        { subscription_name: 'Pro 月度订阅', active_users: 1, average_daily_usage_7d: 2.87, subscription_price: 38.5, sort_order: 20 },
        { subscription_name: '日卡', active_users: 0, average_daily_usage_7d: 0, subscription_price: 3.99, sort_order: 30 },
        { subscription_name: '团队订阅', active_users: 1, average_daily_usage_7d: 6.8, subscription_price: 96, sort_order: 40 }
      ]
      utilizationTrends.value = []
    } else {
      const response = await getSubscriptionAnalytics()
      rows.value = response.items || []
      packageSummaries.value = response.package_summaries || []
      utilizationTrends.value = response.utilization_trends || []
    }
  } catch (err) {
    console.error('Error loading subscription analytics:', err)
    rows.value = []
    packageSummaries.value = []
    utilizationTrends.value = []
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
