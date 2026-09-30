<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-xl font-semibold text-gray-900 dark:text-white">订阅用户分析</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">按订阅用户汇总充值、实际计费与每日额度使用情况。</p>
        </div>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </div>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div v-for="card in summaryCards" :key="card.label" class="card p-4">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ card.label }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ card.value }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ card.hint }}</p>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <div>
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">每日额度利用率</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">按每日平均使用量占日限额比例从高到低排列。</p>
          </div>
          <span class="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ rows.length }} 位订阅用户</span>
        </div>
        <div v-if="loading" class="flex h-72 items-center justify-center"><LoadingSpinner /></div>
        <div v-else-if="error" class="flex h-72 flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400">
          <span>暂时无法加载订阅分析数据</span>
          <button type="button" class="btn btn-secondary" @click="load">重试</button>
        </div>
        <div v-else-if="!rows.length" class="flex h-72 items-center justify-center text-sm text-gray-500 dark:text-gray-400">暂无活跃订阅用户数据</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[860px] text-left text-sm">
            <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="px-4 py-3 font-medium">用户 / 订阅</th>
                <th class="px-4 py-3 text-right font-medium">平均充值</th>
                <th class="px-4 py-3 text-right font-medium">月均实际计费</th>
                <th class="px-4 py-3 text-right font-medium">日均使用</th>
                <th class="px-4 py-3 text-right font-medium">日限额</th>
                <th class="px-4 py-3 font-medium">日限额利用率</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in rows" :key="`${row.subscription_id}-${row.user_id}`" class="hover:bg-gray-50/70 dark:hover:bg-dark-800/60">
                <td class="px-4 py-3">
                  <p class="font-medium text-gray-900 dark:text-white">{{ row.email }}</p>
                  <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ row.subscription_name }}</p>
                </td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.average_recharge_amount) }}</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.average_monthly_actual_cost) }}</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.average_daily_usage) }}</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-gray-200">{{ money(row.daily_limit) }}</td>
                <td class="px-4 py-3">
                  <div class="flex min-w-48 items-center gap-3">
                    <div class="h-2 flex-1 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                      <div class="h-full rounded" :class="utilizationColor(row.daily_limit_utilization)" :style="{ width: `${Math.min(row.daily_limit_utilization, 100)}%` }" />
                    </div>
                    <span class="w-12 text-right text-xs font-medium tabular-nums text-gray-700 dark:text-gray-200">{{ row.daily_limit_utilization.toFixed(1) }}%</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getSubscriptionAnalytics, type SubscriptionAnalyticsItem } from '@/api/admin/dashboard'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const route = useRoute()
const loading = ref(false)
const error = ref(false)
const rows = ref<SubscriptionAnalyticsItem[]>([])

const demoRows: SubscriptionAnalyticsItem[] = [
  { subscription_id: 1, subscription_name: 'Pro 月度订阅', user_id: 101, email: 'alex@example.com', average_recharge_amount: 38.5, average_monthly_actual_cost: 31.2, average_daily_usage: 2.87, daily_limit: 3, daily_limit_utilization: 95.7 },
  { subscription_id: 2, subscription_name: '团队订阅', user_id: 102, email: 'morgan@example.com', average_recharge_amount: 96, average_monthly_actual_cost: 74.4, average_daily_usage: 6.8, daily_limit: 8, daily_limit_utilization: 85 },
  { subscription_id: 3, subscription_name: '基础订阅', user_id: 103, email: 'sam@example.com', average_recharge_amount: 18, average_monthly_actual_cost: 10.8, average_daily_usage: 0.49, daily_limit: 1, daily_limit_utilization: 49 }
]

const summaryCards = computed(() => {
  const count = rows.value.length || 1
  const averageRecharge = rows.value.reduce((sum, row) => sum + row.average_recharge_amount, 0) / count
  const averageBilling = rows.value.reduce((sum, row) => sum + row.average_monthly_actual_cost, 0) / count
  const averageUtilization = rows.value.reduce((sum, row) => sum + row.daily_limit_utilization, 0) / count
  return [
    { label: '订阅用户', value: `${rows.value.length}`, hint: '当前纳入统计的有效订阅' },
    { label: '平均充值费用', value: money(averageRecharge), hint: '按订阅用户的充值记录均值' },
    { label: '平均日限额利用率', value: `${averageUtilization.toFixed(1)}%`, hint: `月均实际计费 ${money(averageBilling)}` }
  ]
})

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
    } else {
      const response = await getSubscriptionAnalytics()
      rows.value = response.items || []
    }
    rows.value.sort((left, right) => right.daily_limit_utilization - left.daily_limit_utilization)
  } catch (err) {
    console.error('Error loading subscription analytics:', err)
    rows.value = []
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
