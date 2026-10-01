import { defineConfig, loadEnv, Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import checker from 'vite-plugin-checker'
import { resolve } from 'path'

const mockUser = {
  id: 1001,
  username: 'mock-user',
  email: 'mock@example.com',
  role: 'admin',
  balance: 256.8,
  concurrency: 5,
  rpm_limit: 0,
  status: 'active',
  allowed_groups: null,
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-09-25T00:00:00Z',
  updated_at: '2026-09-25T00:00:00Z'
}

function mockResponse(res: import('http').ServerResponse, data: unknown, status = 200): void {
  res.writeHead(status, { 'Content-Type': 'application/json' })
  res.end(JSON.stringify({ code: 0, message: 'ok', data }))
}

async function readMockJsonBody(req: import('http').IncomingMessage): Promise<Record<string, unknown>> {
  const chunks: Buffer[] = []
  for await (const chunk of req) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk))
  }
  if (chunks.length === 0) return {}
  return JSON.parse(Buffer.concat(chunks).toString('utf8')) as Record<string, unknown>
}

function createMockApi(): Plugin {
  const mockDashboardStats = {
    total_users: 128,
    today_new_users: 4,
    active_users: 43,
    hourly_active_users: 17,
    stats_updated_at: new Date().toISOString(),
    stats_stale: false,
    total_api_keys: 286,
    active_api_keys: 241,
    total_accounts: 36,
    normal_accounts: 31,
    error_accounts: 2,
    ratelimit_accounts: 2,
    overload_accounts: 1,
    total_requests: 186420,
    total_input_tokens: 102840000,
    total_output_tokens: 28460000,
    total_cache_creation_tokens: 4960000,
    total_cache_read_tokens: 41820000,
    total_tokens: 178080000,
    total_cost: 1162.48,
    total_actual_cost: 1078.63,
    total_account_cost: 823.14,
    today_requests: 1842,
    today_input_tokens: 1084000,
    today_output_tokens: 326000,
    today_cache_creation_tokens: 52000,
    today_cache_read_tokens: 418000,
    today_tokens: 1880000,
    today_cost: 12.84,
    today_actual_cost: 11.92,
    today_account_cost: 8.71,
    average_duration_ms: 742,
    uptime: 2592000,
    rpm: 24,
    tpm: 28800
  }
  const mockDates = Array.from({ length: 30 }, (_, index) => {
    const date = new Date()
    date.setHours(0, 0, 0, 0)
    date.setDate(date.getDate() - (29 - index))
    return date.toISOString().slice(0, 10)
  })
  const formatMockDate = (date: Date): string =>
    `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  const mockTrend = mockDates.map((date, index) => {
    const wave = Math.round(Math.sin(index * 0.72) * 100000)
    const inputTokens = 740000 + index * 11000 + wave
    const outputTokens = 218000 + index * 3700 + Math.round(wave * 0.22)
    const cacheReadTokens = 290000 + index * 5200 + Math.round(wave * 0.3)
    const cacheCreationTokens = 36000 + index * 650
    return {
      date,
      requests: 1180 + index * 22 + (index % 5) * 41,
      input_tokens: inputTokens,
      output_tokens: outputTokens,
      cache_creation_tokens: cacheCreationTokens,
      cache_read_tokens: cacheReadTokens,
      total_tokens: inputTokens + outputTokens + cacheReadTokens + cacheCreationTokens,
      cost: Number((7.8 + index * 0.15 + (index % 4) * 0.24).toFixed(4)),
      actual_cost: Number((7.2 + index * 0.14 + (index % 4) * 0.21).toFixed(4))
    }
  })
  const mockModels = [
    { model: 'gpt-5', requests: 64320, input_tokens: 46200000, output_tokens: 13940000, cache_creation_tokens: 2140000, cache_read_tokens: 18480000, total_tokens: 80760000, cost: 584.62, actual_cost: 541.38, account_cost: 418.21 },
    { model: 'claude-sonnet-4', requests: 42860, input_tokens: 28460000, output_tokens: 8420000, cache_creation_tokens: 1460000, cache_read_tokens: 12620000, total_tokens: 50960000, cost: 348.47, actual_cost: 323.61, account_cost: 247.86 },
    { model: 'gemini-2.5-pro', requests: 29480, input_tokens: 18180000, output_tokens: 4840000, cache_creation_tokens: 820000, cache_read_tokens: 6720000, total_tokens: 30560000, cost: 186.19, actual_cost: 173.46, account_cost: 132.54 }
  ]
  const mockUsers = [
    { user_id: 1001, email: 'mock@example.com', username: 'mock-user', actual_cost: 92.48, requests: 16320, tokens: 21940000 },
    { user_id: 1002, email: 'lin@sub2api.local', username: 'Lin', actual_cost: 74.16, requests: 12640, tokens: 17480000 },
    { user_id: 1003, email: 'chen@sub2api.local', username: 'Chen', actual_cost: 61.73, requests: 10980, tokens: 14860000 },
    { user_id: 1004, email: 'yu@sub2api.local', username: 'Yu', actual_cost: 48.92, requests: 8760, tokens: 11240000 },
    { user_id: 1005, email: 'wu@sub2api.local', username: 'Wu', actual_cost: 39.65, requests: 7440, tokens: 9640000 }
  ]
  const mockUsersTrend = mockDates.flatMap((date, dayIndex) => mockUsers.map((user, userIndex) => ({
    date,
    user_id: user.user_id,
    email: user.email,
    username: user.username,
    requests: Math.round(user.requests / 30) + dayIndex * 2 + userIndex * 7,
    tokens: Math.round(user.tokens / 30) + dayIndex * 4800 + userIndex * 18600,
    cost: Number((user.actual_cost / 30).toFixed(4)),
    actual_cost: Number((user.actual_cost / 30).toFixed(4))
  })))
  const mockGroups = [
    { id: 1, name: 'OpenAI Pro', description: '本地模拟 OpenAI 分组', platform: 'openai', rate_multiplier: 1.2, cache_read_multiplier: 1.1, night_cache_read_multiplier: 1.2, is_exclusive: false, status: 'active', subscription_type: 'subscription', daily_limit_usd: 25, weekly_limit_usd: 100, monthly_limit_usd: 300, long_context_pricing_enabled: true, peak_rate_enabled: false, peak_start: '', peak_end: '', peak_rate_multiplier: 1, night_rate_enabled: true, night_start: '01:30', night_end: '06:30', night_rate_multiplier: 1.5, rpm_limit: 0, model_pricing: [], sort_order: 1 },
    { id: 2, name: 'Claude Standard', description: '本地模拟 Anthropic 分组', platform: 'anthropic', rate_multiplier: 1, cache_read_multiplier: 1.1, night_cache_read_multiplier: 1.2, is_exclusive: false, status: 'active', subscription_type: 'standard', daily_limit_usd: null, weekly_limit_usd: null, monthly_limit_usd: null, long_context_pricing_enabled: true, peak_rate_enabled: false, peak_start: '', peak_end: '', peak_rate_multiplier: 1, night_rate_enabled: true, night_start: '01:30', night_end: '06:30', night_rate_multiplier: 1.5, rpm_limit: 0, model_pricing: [], sort_order: 2 }
  ]
  for (let id = 3; id <= 12; id += 1) {
    const template = mockGroups[(id - 1) % 2]
    mockGroups.push({
      ...template,
      id,
      name: `Mock Group ${id}`,
      sort_order: id,
    })
  }
  const mockSubscriptions = [
    { id: 1, user_id: 1001, group_id: 1, status: 'active', starts_at: '2026-09-01T00:00:00Z', expires_at: '2026-10-01T00:00:00Z', daily_usage_usd: 1.2, weekly_usage_usd: 5.4, monthly_usage_usd: 12.8, daily_window_start: '2026-09-25T00:00:00Z', weekly_window_start: '2026-09-22T00:00:00Z', monthly_window_start: '2026-09-01T00:00:00Z', created_at: '2026-09-01T00:00:00Z', updated_at: '2026-09-25T00:00:00Z', user: mockUser, group: mockGroups[0] }
  ]
  const mockChannels = [{ id: 1, name: 'OpenAI Official', description: '本地模型定价演示渠道', status: 'active', billing_model_source: 'requested', restrict_models: false, group_ids: [1], model_pricing: [], model_mapping: {}, apply_pricing_to_account_stats: true, account_stats_pricing_rules: [], created_at: new Date().toISOString(), updated_at: new Date().toISOString() }]
  const mockDefaultPrices: Record<string, [number, number, number, number]> = {
    // Values are USD per 1M tokens; API responses below convert them to per-token values.
    'gpt-5.5': [5, 30, 5, 0.5], 'gpt-5.6-sol': [5, 30, 6.25, 0.5], 'gpt-5.6-terra': [2, 12, 2.5, 0.2], 'gpt-5.6-luna': [0.2, 1.2, 0.25, 0.02],
    'gpt-6': [10, 50, 12.5, 1], 'gpt-6-astra': [10, 50, 12.5, 1], 'gpt-6-sol': [2, 10, 2.5, 0.2], 'gpt-6-luna': [0.1, 0.5, 0.125, 0.01], 'gpt-6.1-sol': [2, 10, 2.5, 0.1],
    'claude-opus-5-5': [4, 20, 5, 0.2], 'claude-opus-5': [5, 25, 6.25, 0.5], 'claude-sonnet-5-5': [2, 10, 2.5, 0.2], 'claude-sonnet-5': [2, 10, 2.5, 0.2],
    'claude-opus-4-8': [5, 25, 6.25, 0.5], 'claude-opus-4-7': [5, 25, 6.25, 0.5], 'claude-opus-4-6': [5, 25, 6.25, 0.5], 'claude-sonnet-4-6': [3, 15, 3.75, 0.3], 'claude-haiku-4-5': [1, 5, 1.25, 0.1]
  }
  return {
    name: 'local-mock-api',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use(async (req, res, next) => {
        const path = new URL(req.url || '/', 'http://localhost').pathname
        if (!path.startsWith('/api/v1/')) return next()

        if (path === '/api/v1/auth/login' && req.method === 'POST') {
          return mockResponse(res, {
            access_token: 'mock-access-token',
            refresh_token: 'mock-refresh-token',
            expires_in: 86400,
            token_type: 'Bearer',
            user: mockUser
          })
        }
        if (path === '/api/v1/auth/me') return mockResponse(res, { ...mockUser, run_mode: 'standard' })
        if (path === '/api/v1/auth/refresh') {
          return mockResponse(res, { access_token: 'mock-access-token', refresh_token: 'mock-refresh-token', expires_in: 86400, token_type: 'Bearer' })
        }
        if (path === '/api/v1/auth/logout') return mockResponse(res, {})
        if (path === '/api/v1/settings/public') {
          return mockResponse(res, { site_name: 'Sub2API Local Mock', registration_enabled: true, email_login_enabled: true, model_plaza_enabled: true, model_plaza_require_auth: false, model_plaza_description: '本地模型价格展示，价格与管理员模型定价配置同步。' })
        }
        if (path === '/api/v1/model-plaza') {
          const modelNames: Array<[string, string]> = [['gpt-5.5', 'openai'], ['gpt-5.6-sol', 'openai'], ['gpt-5.6-terra', 'openai'], ['gpt-5.6-luna', 'openai'], ['gpt-6-astra', 'openai'], ['gpt-6-sol', 'openai'], ['gpt-6-luna', 'openai'], ['gpt-6.1-sol', 'openai'], ['claude-opus-5-5', 'anthropic'], ['claude-opus-5', 'anthropic'], ['claude-sonnet-5-5', 'anthropic'], ['claude-sonnet-5', 'anthropic'], ['claude-opus-4-8', 'anthropic'], ['claude-opus-4-7', 'anthropic'], ['claude-opus-4-6', 'anthropic'], ['claude-sonnet-4-6', 'anthropic'], ['claude-haiku-4-5', 'anthropic']]
          const toPricing = (model: string, platform: string) => {
            const rule = mockChannels[0].model_pricing.find((item) => item.models.includes(model))
            const defaults = mockDefaultPrices[model] || [0, 0, 0, 0]
            const input = rule?.input_price ?? defaults[0] / 1_000_000
            const output = rule?.output_price ?? defaults[1] / 1_000_000
            const cacheWrite = rule?.cache_write_price ?? defaults[2] / 1_000_000
            const cacheRead = rule?.cache_read_price ?? defaults[3] / 1_000_000
            return { billing_mode: 'token', input_price: input, output_price: output, cache_write_price: cacheWrite, cache_read_price: cacheRead, image_input_price: null, image_output_price: null, per_request_price: null, intervals: [] }
          }
          const makeGroup = (id: number, name: string, platform: string) => ({
            id, name, description: `${name} · 管理员定价实时同步`, platform, subscription_type: 'standard', rate_multiplier: 1,
            peak_rate_enabled: false, peak_start: '', peak_end: '', peak_rate_multiplier: 1, is_exclusive: false,
            image_rate_independent: false, image_rate_multiplier: 1, long_context_pricing_enabled: true,
            models: modelNames.filter(([, itemPlatform]) => itemPlatform === platform).map(([model]) => {
              const p = mockDefaultPrices[model] || [0, 0, 0, 0]
              return { name: model, platform, pricing: toPricing(model, platform), official_pricing: { input_price: p[0] / 1_000_000, output_price: p[1] / 1_000_000, cache_write_price: p[2] / 1_000_000, cache_read_price: p[3] / 1_000_000 } }
            })
          })
          return mockResponse(res, { description: '本地模型价格展示，价格与管理员模型定价配置同步。', groups: [makeGroup(1, 'OpenAI GPT 模型', 'openai'), makeGroup(2, 'Anthropic Claude 模型', 'anthropic')] })
        }
        if (path === '/api/v1/groups/available') return mockResponse(res, mockGroups)
        if (path === '/api/v1/groups/rates') return mockResponse(res, {})
        if (path === '/api/v1/usage/stats') {
          return mockResponse(res, {
            total_requests: 1284,
            total_input_tokens: 860000,
            total_output_tokens: 224000,
            total_cache_creation_tokens: 18000,
            total_cache_read_tokens: 105600,
            total_cache_tokens: 123600,
            total_tokens: 1207600,
            total_cost: 3.8079,
            total_actual_cost: 3.8079,
            average_duration_ms: 820,
            endpoints: []
          })
        }
        if (path === '/api/v1/usage/dashboard/stats') {
          return mockResponse(res, {
            total_api_keys: 2, active_api_keys: 2, total_requests: 1284,
            total_input_tokens: 860000, total_output_tokens: 224000, total_cache_creation_tokens: 18000, total_cache_read_tokens: 105600,
            total_cache_tokens: 123600, total_tokens: 1207600, total_cost: 3.8079, total_actual_cost: 3.8079, total_account_cost: 0,
            today_requests: 36, today_input_tokens: 24000, today_output_tokens: 6800, today_cache_creation_tokens: 600, today_cache_read_tokens: 4620,
            today_cache_tokens: 5220, today_tokens: 36620, today_cost: 0.12, today_actual_cost: 0.12, today_account_cost: 0,
            average_duration_ms: 820, rpm: 1, tpm: 460
          })
        }
        if (path === '/api/v1/usage/dashboard/trend') {
          return mockResponse(res, { trend: Array.from({ length: 7 }, (_, index) => ({
            date: `2026-09-${String(19 + index).padStart(2, '0')}`,
            requests: 80 + index * 13,
            input_tokens: 36000 + index * 3200,
            output_tokens: 12000 + index * 1100,
            cache_creation_tokens: 1800 + index * 160,
            cache_read_tokens: 14200 + index * 940,
            total_tokens: 62000 + index * 5400,
            cost: 0.22 + index * 0.03,
            actual_cost: 0.22 + index * 0.03
          })), start_date: '2026-09-19', end_date: '2026-09-25', granularity: 'day' })
        }
        if (path === '/api/v1/usage/dashboard/models') {
          return mockResponse(res, { models: [
            { model: 'gpt-5', requests: 860, input_tokens: 540000, output_tokens: 164000, total_tokens: 704000, actual_cost: 2.46, cost: 2.46 },
            { model: 'claude-sonnet-4', requests: 424, input_tokens: 320000, output_tokens: 60000, total_tokens: 380000, actual_cost: 1.35, cost: 1.35 }
          ], start_date: '2026-09-19', end_date: '2026-09-25' })
        }
        if (path === '/api/v1/usage/dashboard/snapshot-v2') {
          const params = new URL(req.url || '/', 'http://localhost').searchParams
          const requestedGranularity = params.get('granularity')
          const snapshotGranularity = requestedGranularity === 'hour' ? 'hour' : 'day'
          const start = params.get('start_date') || mockDates[0]
          const end = params.get('end_date') || mockDates[mockDates.length - 1]
          const startTime = new Date(`${start}T00:00:00`)
          const endTime = new Date(`${end}T00:00:00`)
          const rangeDays = Math.max(1, Math.ceil((endTime.getTime() - startTime.getTime()) / 86400000))
          const pointCount = snapshotGranularity === 'hour' ? 24 : Math.min(rangeDays, 30)
          const trend = Array.from({ length: pointCount }, (_, index) => {
            const pointDate = new Date(startTime)
            if (snapshotGranularity === 'hour') {
              pointDate.setHours(pointDate.getHours() + index)
            } else {
              pointDate.setDate(pointDate.getDate() + index)
            }
            const date = snapshotGranularity === 'hour'
              ? `${formatMockDate(pointDate)} ${String(pointDate.getHours()).padStart(2, '0')}:00`
              : formatMockDate(pointDate)
            const wave = Math.round(Math.sin(index * 0.72) * 2800)
            const inputTokens = 36000 + index * 2100 + wave
            const outputTokens = 12000 + index * 760 + Math.round(wave * 0.2)
            const cacheCreationTokens = 1800 + index * 90
            const cacheReadTokens = 14200 + index * 520 + Math.round(wave * 0.35)
            const weightedCacheReadTokens = Math.round(cacheReadTokens * 1.1)
            return {
              date,
              requests: 80 + index * 7,
              input_tokens: inputTokens,
              output_tokens: outputTokens,
              cache_creation_tokens: cacheCreationTokens,
              cache_read_tokens: weightedCacheReadTokens,
              total_tokens: inputTokens + outputTokens + cacheCreationTokens + weightedCacheReadTokens,
              cost: Number((0.22 + index * 0.012).toFixed(4)),
              actual_cost: Number((0.22 + index * 0.012).toFixed(4))
            }
          })
          return mockResponse(res, {
            generated_at: new Date().toISOString(),
            start_date: start,
            end_date: end,
            granularity: snapshotGranularity,
            trend,
            groups: [
              { group_id: 1, group_name: 'OpenAI Pro', requests: 820, total_tokens: 748000, cost: 2.46, actual_cost: 2.46 },
              { group_id: 2, group_name: 'Claude Standard', requests: 464, total_tokens: 459600, cost: 1.35, actual_cost: 1.35 }
            ]
          })
        }
        if (path === '/api/v1/usage') {
          return mockResponse(res, { items: [], total: 0, page: 1, page_size: 20, pages: 0 })
        }
        if (path === '/api/v1/admin/dashboard/stats') return mockResponse(res, mockDashboardStats)
        if (path === '/api/v1/admin/channels' && req.method === 'GET') return mockResponse(res, { items: mockChannels, total: mockChannels.length })
        if (path === '/api/v1/admin/channels/model-pricing') {
          const model = new URL(req.url || '/', 'http://localhost').searchParams.get('model') || ''
          const prices = mockDefaultPrices[model]
          return mockResponse(res, prices ? { found: true, input_price: prices[0] / 1_000_000, output_price: prices[1] / 1_000_000, cache_write_price: prices[2] / 1_000_000, cache_read_price: prices[3] / 1_000_000 } : { found: false })
        }
        if (/^\/api\/v1\/admin\/channels\/\d+$/.test(path) && req.method === 'PUT') {
          const id = Number(path.split('/').pop()); const body = await readMockJsonBody(req); const channel = mockChannels.find(item => item.id === id)
          if (!channel) return mockResponse(res, { message: 'not found' }, 404)
          Object.assign(channel, body, { updated_at: new Date().toISOString() }); return mockResponse(res, channel)
        }
        if (path === '/api/v1/admin/dashboard/snapshot-v2') {
          return mockResponse(res, {
            generated_at: new Date().toISOString(),
            start_date: mockDates[0],
            end_date: mockDates[mockDates.length - 1],
            granularity: 'day',
            stats: mockDashboardStats,
            trend: mockTrend,
            models: mockModels,
            groups: [],
            users_trend: mockUsersTrend
          })
        }
        if (path === '/api/v1/admin/dashboard/trend') {
          return mockResponse(res, {
            trend: mockTrend,
            start_date: mockDates[0],
            end_date: mockDates[mockDates.length - 1],
            granularity: 'day'
          })
        }
        if (path === '/api/v1/admin/dashboard/models') {
          return mockResponse(res, { models: mockModels, start_date: mockDates[0], end_date: mockDates[mockDates.length - 1] })
        }
        if (path === '/api/v1/admin/dashboard/users-trend') {
          return mockResponse(res, {
            trend: mockUsersTrend,
            start_date: mockDates[0],
            end_date: mockDates[mockDates.length - 1],
            granularity: 'day'
          })
        }
        if (path === '/api/v1/admin/dashboard/users-ranking') {
          return mockResponse(res, {
            ranking: mockUsers,
            total_actual_cost: mockUsers.reduce((total, user) => total + user.actual_cost, 0),
            total_requests: mockUsers.reduce((total, user) => total + user.requests, 0),
            total_tokens: mockUsers.reduce((total, user) => total + user.tokens, 0),
            start_date: mockDates[0],
            end_date: mockDates[mockDates.length - 1]
          })
        }
        if (path === '/api/v1/admin/dashboard/subscription-analytics') {
          const items = mockUsers.map((user, index) => ({
            subscription_id: index + 1,
            subscription_name: index % 2 === 0 ? 'OpenAI Pro' : 'Claude Standard',
            user_id: user.user_id,
            email: user.email,
            average_recharge_amount: 58 + index * 14.5,
            average_monthly_actual_cost: user.actual_cost,
            average_daily_usage: Number((user.actual_cost / 30).toFixed(2)),
            daily_limit: index % 2 === 0 ? 4 : 5,
            daily_limit_utilization: Number(((user.actual_cost / 30 / (index % 2 === 0 ? 4 : 5)) * 100).toFixed(1))
          })).sort((left, right) => right.daily_limit_utilization - left.daily_limit_utilization)
          const packages = new Map<string, { users: Set<number>; usage: number }>()
          items.forEach((item) => {
            const packageItem = packages.get(item.subscription_name) || { users: new Set<number>(), usage: 0 }
            packageItem.users.add(item.user_id)
            packageItem.usage += item.average_daily_usage
            packages.set(item.subscription_name, packageItem)
          })
          return mockResponse(res, {
            generated_at: new Date().toISOString(),
            items,
            package_summaries: Array.from(packages, ([subscription_name, packageItem]) => ({
              subscription_name,
              active_users: packageItem.users.size,
              average_daily_usage_7d: Number((packageItem.usage / packageItem.users.size).toFixed(2)),
              sort_order: subscription_name === 'OpenAI Pro' ? 1 : 2
            })),
            utilization_trends: items.map((item, subscriptionIndex) => {
              const baseline = item.daily_limit_utilization
              return {
                subscription_id: item.subscription_id,
                subscription_name: item.subscription_name,
                period_start: mockDates[0],
                period_end: mockDates[mockDates.length - 1],
                daily_limit: item.daily_limit,
                trend: mockDates.map((date, dayIndex) => ({
                  date,
                  utilization: Number(Math.max(0, baseline + Math.sin(dayIndex * 0.65 + subscriptionIndex) * 9 + (dayIndex % 4) * 1.5).toFixed(1))
                }))
              }
            })
          })
        }
        if (path === '/api/v1/admin/groups/billing-settings' && req.method === 'PUT') {
          const settings = await readMockJsonBody(req)
          if (settings.kind === 'night') {
            mockGroups.forEach((group) => {
              group.night_rate_enabled = Boolean(settings.night_rate_enabled)
              group.night_start = String(settings.night_start ?? '')
              group.night_end = String(settings.night_end ?? '')
              group.night_rate_multiplier = Number(settings.night_rate_multiplier)
              group.night_cache_read_multiplier = Number(settings.night_cache_read_multiplier)
            })
          }
          return mockResponse(res, { updated: mockGroups.length })
        }
        if (path === '/api/v1/admin/groups/all') return mockResponse(res, mockGroups)
        if (path === '/api/v1/admin/groups') return mockResponse(res, { items: mockGroups, total: mockGroups.length, page: 1, page_size: 20, pages: 1 })
        if (path === '/api/v1/admin/subscriptions') return mockResponse(res, { items: mockSubscriptions, total: mockSubscriptions.length, page: 1, page_size: 20, pages: 1 })
        if (path === '/api/v1/admin/users') return mockResponse(res, { items: [mockUser], total: 1, page: 1, page_size: 20, pages: 1 })
        if (path === '/api/v1/admin/groups/usage-summary' || path === '/api/v1/admin/groups/capacity-summary') return mockResponse(res, [])
        if (path === '/api/v1/announcements') return mockResponse(res, [])
        return mockResponse(res, { items: [], total: 0, page: 1, page_size: 20, pages: 0 })
      })
    }
  }
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (character) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[character] || character)
}

function isSafeImageUrl(value: string): boolean {
  const trimmed = value.trim()
  if ((trimmed.startsWith('/') && !trimmed.startsWith('//')) || /^data:image\//i.test(trimmed)) {
    return true
  }
  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function injectBranding(html: string, config: { site_name?: string; site_logo?: string }): string {
  let brandedHtml = html
  const siteName = config.site_name?.trim()
  if (siteName) {
    brandedHtml = brandedHtml.replace(
      /<title>[^<]*<\/title>/i,
      `<title>${escapeHtml(siteName)} - AI API Gateway</title>`,
    )
  }

  const siteLogo = config.site_logo?.trim()
  if (siteLogo && isSafeImageUrl(siteLogo)) {
    brandedHtml = brandedHtml.replace(
      /<link\s+rel=["']icon["'][^>]*>/i,
      `<link rel="icon" href="${escapeHtml(siteLogo)}" />`,
    )
  }
  return brandedHtml
}

/**
 * Vite 插件：开发模式下注入公开配置到 index.html
 * 与生产模式的后端注入行为保持一致，消除闪烁
 */
function injectPublicSettings(backendUrl: string): Plugin {
  return {
    name: 'inject-public-settings',
    apply: 'serve',
    transformIndexHtml: {
      order: 'pre',
      async handler(html) {
        try {
          const response = await fetch(`${backendUrl}/api/v1/settings/public`, {
            signal: AbortSignal.timeout(2000)
          })
          if (response.ok) {
            const data = await response.json()
            if (data.code === 0 && data.data) {
              const script = `<script>window.__APP_CONFIG__=${JSON.stringify(data.data)};</script>`
              return injectBranding(html, data.data).replace('</head>', `${script}\n</head>`)
            }
          }
        } catch (e) {
          console.warn('[vite] 无法获取公开配置，将回退到 API 调用:', (e as Error).message)
        }
        return html
      }
    }
  }
}

export default defineConfig(({ mode }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_DEV_PROXY_TARGET || 'http://localhost:8080'
  const devPort = Number(env.VITE_DEV_PORT || 3000)

  return {
    plugins: [
      vue(),
      checker({
        vueTsc: true
      }),
      ...(env.VITE_MOCK_API === 'true' ? [createMockApi()] : []),
      injectPublicSettings(backendUrl)
    ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      // 使用 vue-i18n 运行时版本，避免 CSP unsafe-eval 问题
      'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js'
    }
  },
  define: {
    // 启用 vue-i18n JIT 编译，在 CSP 环境下处理消息插值
    // JIT 编译器生成 AST 对象而非 JS 代码，无需 unsafe-eval
    __INTLIFY_JIT_COMPILATION__: true
  },
  build: {
    outDir: '../backend/internal/web/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        /**
         * 手动分包配置
         * 分离第三方库并按功能合并应用代码，避免循环依赖
         */
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            // Vue 核心库
            if (
              id.includes('/vue/') ||
              id.includes('/vue-router/') ||
              id.includes('/pinia/') ||
              id.includes('/@vue/')
            ) {
              return 'vendor-vue'
            }

            // UI 工具库（较大，单独分离）
            if (id.includes('/@vueuse/') || id.includes('/xlsx/')) {
              return 'vendor-ui'
            }

            // 图表库
            if (id.includes('/chart.js/') || id.includes('/vue-chartjs/')) {
              return 'vendor-chart'
            }

            // 国际化
            if (id.includes('/vue-i18n/') || id.includes('/@intlify/')) {
              return 'vendor-i18n'
            }

            // Stripe 仅在支付流程中按需加载，避免进入首页公共依赖。
            if (id.includes('/@stripe/stripe-js/')) {
              return 'vendor-stripe'
            }

            // 其他小型第三方库合并
            return 'vendor-misc'
          }

          // 应用代码：按入口点自动分包，不手动干预
          // 这样可以避免循环依赖，同时保持合理的 chunk 数量
        }
      }
    }
  },
    server: {
      host: '0.0.0.0',
      port: devPort,
      proxy: env.VITE_MOCK_API === 'true' ? undefined : {
        '/api': {
          target: backendUrl,
          changeOrigin: true
        },
        '/v1': {
          target: backendUrl,
          changeOrigin: true
        },
        '/setup': {
          target: backendUrl,
          changeOrigin: true
        }
      }
    }
  }
})
