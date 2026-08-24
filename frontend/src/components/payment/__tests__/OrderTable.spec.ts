import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { PaymentOrder } from '@/types/payment'
import OrderTable from '../OrderTable.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const DataTableStub = {
  props: ['data'],
  template: '<div><slot name="cell-user_email" :value="data[0]?.user_email" :row="data[0]" /></div>',
}

describe('OrderTable', () => {
  it('shows a paid order user registration promo group and code tooltip', () => {
    const order: PaymentOrder = {
      id: 1,
      user_id: 2,
      amount: 100,
      pay_amount: 100,
      fee_rate: 0,
      payment_type: 'alipay',
      out_trade_no: 'ORDER-1',
      status: 'PAID',
      order_type: 'balance',
      created_at: '2026-08-01T00:00:00Z',
      expires_at: '2026-08-01T01:00:00Z',
      refund_amount: 0,
      user_email: 'user@example.com',
      registration_promo_group: '第一组',
      registration_promo_code: 'EARLYCODE',
    }

    const wrapper = mount(OrderTable, {
      props: { orders: [order], loading: false, showUser: true },
      global: { stubs: { DataTable: DataTableStub } },
    })

    expect(wrapper.get('[title="EARLYCODE"]').text()).toBe('(第一组)')
  })
})
