<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header with Day Switcher -->
      <div class="flex flex-wrap items-center justify-end gap-3">
        <div class="flex items-center gap-2">
          <input v-model="customStartDate" type="date" class="input w-36" :title="t('payment.admin.customStartDate')" />
          <span class="text-sm text-gray-400">—</span>
          <input v-model="customEndDate" type="date" class="input w-36" :title="t('payment.admin.customEndDate')" />
          <button type="button" class="btn btn-primary" :disabled="loading || !customStartDate || !customEndDate" @click="applyCustomRange">
            {{ t('payment.admin.applyDateRange') }}
          </button>
          <button v-if="customActive" type="button" class="btn btn-secondary" :disabled="loading" @click="clearCustomRange">
            {{ t('payment.admin.clearDateRange') }}
          </button>
        </div>
        <div class="flex items-center gap-2">
          <div class="flex rounded-lg border border-gray-200 dark:border-dark-600">
            <button
              v-for="d in DAYS_OPTIONS"
              :key="d"
              type="button"
              class="px-3 py-1.5 text-xs font-medium transition-colors first:rounded-l-lg last:rounded-r-lg"
              :class="days === d
                ? 'bg-primary-600 text-white'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
              @click="selectDays(d)"
            >
              {{ d }}{{ t('payment.admin.daySuffix') }}
            </button>
          </div>
          <button @click="loadDashboard" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <!-- Dashboard Content -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.paymentDistribution') }}</h3>
            <div v-if="!stats.payment_methods?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
            <div v-else class="space-y-3">
              <div v-for="method in stats.payment_methods" :key="method.type" class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span :class="['inline-block h-3 w-3 rounded-full', methodColor(method.type)]"></span>
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + method.type, method.type) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">&yen;{{ formatMoney(method.amount) }}</span>
                  <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">({{ method.count }})</span>
                </div>
              </div>
            </div>
          </div>
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.topUsers') }}</h3>
            <div v-if="!topUsers.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
            <div v-else class="space-y-2">
              <div v-for="(user, idx) in topUsers" :key="user.user_id" class="flex items-center justify-between rounded-lg px-3 py-2 hover:bg-gray-50 dark:hover:bg-dark-700">
                <div class="flex items-center gap-3">
                  <span :class="['flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold', rankClass(idx)]">{{ idx + 1 }}</span>
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ user.email }}</span>
                </div>
                <span class="text-sm font-medium text-gray-900 dark:text-white">&yen;{{ user.amount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const customStartDate = ref('')
const customEndDate = ref('')
const customActive = ref(false)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: 'bg-blue-500', wxpay: 'bg-green-500',
    alipay_direct: 'bg-blue-400', wxpay_direct: 'bg-green-400',
    stripe: 'bg-purple-500',
  }
  return c[type] || 'bg-gray-400'
}

function rankClass(idx: number): string {
  if (idx === 0) return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  if (idx === 1) return 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
  if (idx === 2) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

async function loadDashboard() {
  loading.value = true
  try {
    const params = customActive.value
      ? { start_date: customStartDate.value, end_date: customEndDate.value, timezone: Intl.DateTimeFormat().resolvedOptions().timeZone }
      : { days: days.value }
    const res = await adminPaymentAPI.getDashboard(params)
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function selectDays(value: number) {
  customActive.value = false
  days.value = value
}

function applyCustomRange() {
  if (!customStartDate.value || !customEndDate.value) return
  if (customEndDate.value < customStartDate.value) {
    appStore.showError(t('payment.admin.exportDateRangeInvalid'))
    return
  }
  customActive.value = true
  loadDashboard()
}

function clearCustomRange() {
  customStartDate.value = ''
  customEndDate.value = ''
  customActive.value = false
  loadDashboard()
}

function amountOf(value: number | Record<string, number>, currency = 'CNY'): number {
  if (typeof value === 'number') return value
  return value[currency] ?? Object.values(value)[0] ?? 0
}
function formatMoney(value: number | Record<string, number>): string { return amountOf(value).toFixed(2) }
const topUsers = computed(() => {
  const value = stats.value?.top_users
  if (!value) return []
  return Array.isArray(value) ? value : (value.CNY ?? Object.values(value)[0] ?? [])
})

watch(days, () => { if (!customActive.value) loadDashboard() })
onMounted(() => loadDashboard())
</script>
