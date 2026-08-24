import { flushPromises, shallowMount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminOrdersView from '../AdminOrdersView.vue'

const { showError, getOrders, exportOrders } = vi.hoisted(() => ({
  showError: vi.fn(),
  getOrders: vi.fn(),
  exportOrders: vi.fn(),
}))

vi.mock('vue-i18n', async importOriginal => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess: vi.fn() }),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: { getOrders, exportOrders },
  default: { getOrders, exportOrders },
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'error',
}))

vi.mock('@/components/payment/orderUtils', () => ({
  formatOrderDateTime: (value: string) => value,
}))

function mountView() {
  return shallowMount(AdminOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
      },
    },
  })
}

describe('AdminOrdersView export', () => {
  beforeEach(() => {
    getOrders.mockResolvedValue({ data: { items: [], total: 0 } })
    exportOrders.mockResolvedValue({ data: '充值时间,充值金额' })
    showError.mockReset()
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:payment-orders'),
      revokeObjectURL: vi.fn(),
    })
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('rejects an inverted export date range without sending a request', async () => {
    const wrapper = mountView()
    await flushPromises()

    const dateInputs = wrapper.findAll('input[type="date"]')
    await dateInputs[0].setValue('2026-08-02')
    await dateInputs[1].setValue('2026-08-01')
    await wrapper.get('button[title="payment.admin.exportOrders"]').trigger('click')

    expect(showError).toHaveBeenCalledWith('payment.admin.exportDateRangeInvalid')
    expect(exportOrders).not.toHaveBeenCalled()
  })

  it('downloads successfully paid orders for the selected date range', async () => {
    const wrapper = mountView()
    await flushPromises()

    const dateInputs = wrapper.findAll('input[type="date"]')
    await dateInputs[0].setValue('2026-08-01')
    await dateInputs[1].setValue('2026-08-31')
    await wrapper.get('button[title="payment.admin.exportOrders"]').trigger('click')
    await flushPromises()

    expect(exportOrders).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-08-01',
      end_date: '2026-08-31',
    }))
    expect(URL.createObjectURL).toHaveBeenCalledOnce()
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledOnce()
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:payment-orders')
  })
})
