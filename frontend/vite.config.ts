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
          return mockResponse(res, { site_name: 'Sub2API Local Mock', registration_enabled: true, email_login_enabled: true })
        }
        if (path === '/api/v1/usage/dashboard/stats') {
          return mockResponse(res, {
            total_api_keys: 2, active_api_keys: 2, total_requests: 1284,
            total_input_tokens: 860000, total_output_tokens: 224000, total_cache_creation_tokens: 18000, total_cache_read_tokens: 96000,
            total_tokens: 1198000, total_cost: 3.8079, total_actual_cost: 3.8079, total_account_cost: 0,
            today_requests: 36, today_input_tokens: 24000, today_output_tokens: 6800, today_cache_creation_tokens: 600, today_cache_read_tokens: 4200,
            today_tokens: 35600, today_cost: 0.12, today_actual_cost: 0.12, today_account_cost: 0,
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
        if (path === '/api/v1/usage') {
          return mockResponse(res, { items: [], total: 0, page: 1, page_size: 20, pages: 0 })
        }
        if (path === '/api/v1/admin/dashboard/stats') {
          return mockResponse(res, { total_users: 12, active_users: 8, total_api_keys: 18, active_api_keys: 14, today_requests: 128, total_requests: 12684, today_actual_cost: 3.2, total_actual_cost: 438.9, total_account_cost: 0, today_tokens: 92000, total_tokens: 9840000 })
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
