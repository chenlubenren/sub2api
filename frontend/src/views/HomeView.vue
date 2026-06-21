<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-[#f8f8f8] text-[#101010]"
  >
    <component :is="PixelLoadingOverlay" :visible="showIntroLoading" />

    <div class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 bg-[linear-gradient(rgba(16,16,16,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(16,16,16,0.045)_1px,transparent_1px)] bg-[size:34px_34px]"></div>
      <div class="absolute inset-x-0 top-0 h-5 bg-[#101010]"></div>
    </div>

    <header class="relative z-20 border-b border-[#d6d6d6] bg-[#101010] px-5 py-2 text-white">
      <nav class="mx-auto flex max-w-7xl items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-14 w-14 shrink-0 items-center justify-center overflow-hidden bg-transparent">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="min-w-0">
            <div class="truncate text-base font-semibold text-white sm:text-lg">
              {{ siteName }}
            </div>
            <div class="truncate text-xs text-[#d4a533]">
              {{ siteSubtitle }}
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 sm:gap-3">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="border-2 border-white/70 bg-white/10 p-2 text-white transition-colors hover:border-[#d4a533] hover:text-[#d4a533]"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-2 border-2 border-[#d4a533] bg-[#d4a533] px-4 py-2 text-sm font-semibold text-[#101010] transition-colors hover:bg-[#e8c263]"
          >
            <span class="flex h-6 w-6 items-center justify-center border-2 border-[#101010] bg-[#2d7d46] text-[11px] font-semibold text-white">
              {{ userInitial }}
            </span>
            <span>{{ t('home.dashboard') }}</span>
            <Icon name="arrowRight" size="sm" />
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center gap-2 border-2 border-[#d4a533] bg-[#d4a533] px-4 py-2 text-sm font-semibold text-[#101010] transition-colors hover:bg-[#e8c263]"
          >
            <Icon name="login" size="sm" />
            <span>{{ t('home.login') }}</span>
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 mx-auto w-full max-w-7xl flex-1 px-5 pb-10 pt-8 sm:pt-12 lg:px-6 lg:pt-14">
      <section class="grid items-stretch gap-8 lg:grid-cols-[minmax(0,1.08fr)_minmax(360px,440px)] lg:gap-10">
        <div class="flex h-full flex-col gap-6">
          <div class="inline-flex items-center gap-2 self-start border-2 border-[#101010] bg-white px-4 py-2 text-sm font-semibold text-[#101010] shadow-[4px_4px_0_#101010]">
            <Icon name="shield" size="sm" />
            <span>官方直连</span>
          </div>

          <div class="space-y-4">
            <h1 class="max-w-3xl text-4xl font-black leading-tight tracking-normal text-[#101010] sm:text-5xl lg:text-6xl">
              助研算力供应中心
            </h1>
            <p class="max-w-2xl text-lg leading-8 text-[#101010]/70 sm:text-xl">
              不做阉割替代，不碰多余数据流转，用更稳的官方链路把 Codex 的能力直接带到你的工作流里。
            </p>
          </div>

          <div class="flex flex-wrap gap-3">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="hero-cta hero-cta-primary inline-flex items-center gap-2 border-2 border-[#101010] bg-[#3a5ba0] px-5 py-3 text-sm font-semibold text-white shadow-[4px_4px_0_#101010] transition-colors hover:bg-[#2d7d46]"
            >
              <Icon name="arrowRight" size="sm" class="hero-cta-icon" />
              <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-2 border-2 border-[#101010] bg-white px-5 py-3 text-sm font-semibold text-[#101010] shadow-[4px_4px_0_#101010] transition-colors hover:bg-[#d4a533]"
            >
              <Icon name="book" size="sm" />
              <span>{{ t('home.docs') }}</span>
            </a>
          </div>

          <div class="grid gap-3 md:mt-auto md:grid-cols-3">
            <div class="flex h-full flex-col border-[3px] border-[#101010] bg-white p-5 shadow-[6px_6px_0_#101010]">
              <div class="mb-3 flex h-11 w-11 items-center justify-center border-[3px] border-[#101010] bg-[#2d7d46] text-white shadow-[3px_3px_0_#101010]">
                <Icon name="shield" size="md" />
              </div>
              <h2 class="text-lg font-black leading-7 text-[#101010]">官方模型，不掺水！</h2>
              <p class="mt-2 text-sm leading-6 text-[#101010]/74">
                后台采用真实的官方 PLUS 账号与 PRO 账号提供 Token 供应，不走阉割替代方案，模型能力、可用范围和整体体验都保持官方原生水准。
              </p>
            </div>
            <div class="flex h-full flex-col border-[3px] border-[#101010] bg-white p-5 shadow-[6px_6px_0_#101010]">
              <div class="mb-3 flex h-11 w-11 items-center justify-center border-[3px] border-[#101010] bg-[#3a5ba0] text-white shadow-[3px_3px_0_#101010]">
                <Icon name="lock" size="md" />
              </div>
              <h2 class="text-lg font-black leading-7 text-[#101010]">高度安全，不转卖数据</h2>
              <p class="mt-2 text-sm leading-6 text-[#101010]/74">
                不操控模型输出，不做数据转卖和额外利用，尽量减少中间环节对请求内容的触达，既保护你的对话隐私，也降低本地电脑和工作环境的潜在风险。
              </p>
            </div>
            <div class="flex h-full flex-col border-[3px] border-[#101010] bg-white p-5 shadow-[6px_6px_0_#101010]">
              <div class="mb-3 flex h-11 w-11 items-center justify-center border-[3px] border-[#101010] bg-[#d4a533] text-[#101010] shadow-[3px_3px_0_#101010]">
                <Icon name="sparkles" size="md" />
              </div>
              <h2 class="text-lg font-black leading-7 text-[#101010]">超低价，低于一折使用 token</h2>
              <p class="mt-2 text-sm leading-6 text-[#101010]/74">
                充值按 1 人民币 = 1 美金计算，再叠加专属线路折扣，综合下来约为官方价格的 7% 左右，让高频调用和长期使用都能把预算压得更稳。
              </p>
            </div>
          </div>
        </div>

        <div class="terminal-container mt-2 flex h-full items-end lg:mt-0">
          <div class="terminal-window border-4 border-[#101010] bg-[#101828] shadow-[8px_8px_0_#101010]">
            <div class="terminal-header border-b-2 border-white/10 bg-[#101010]">
              <div class="terminal-buttons">
                <span class="btn-close"></span>
                <span class="btn-minimize"></span>
                <span class="btn-maximize"></span>
              </div>
              <div class="terminal-title">token4research.cn / live route</div>
            </div>

            <div class="terminal-body">
              <div class="code-line line-1">
                <span class="code-prompt">$</span>
                <span class="code-cmd">curl</span>
                <span class="code-flag">-X POST</span>
                <span class="code-url">/v1/messages</span>
              </div>
              <div class="code-line line-2">
                <span class="code-comment"># Routing to upstream...</span>
              </div>
              <div class="code-line line-3">
                <span class="code-success">200 OK</span>
                <span class="code-response">{ "content": "Hello!" }</span>
              </div>
              <div class="code-line line-4">
                <span class="code-prompt">$</span>
                <span class="cursor"></span>
              </div>
            </div>

            <div class="terminal-footer border-t-2 border-white/10 bg-[#111827]/90 px-5 py-4">
              <div class="grid gap-3 sm:grid-cols-3">
                <div class="border-2 border-[#2d7d46] bg-[#2d7d46]/10 px-3 py-2">
                  <div class="terminal-metric-label">Status</div>
                  <div class="terminal-metric-value text-[#7df0a8]">Healthy</div>
                </div>
                <div class="border-2 border-[#3a5ba0] bg-[#3a5ba0]/10 px-3 py-2">
                  <div class="terminal-metric-label">Latency</div>
                  <div class="terminal-metric-value text-[#89a9ff]">128ms</div>
                </div>
                <div class="border-2 border-[#d4a533] bg-[#d4a533]/10 px-3 py-2">
                  <div class="terminal-metric-label">Mode</div>
                  <div class="terminal-metric-value text-[#ffd978]">Codex Ready</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 border-t-4 border-[#a83232] bg-[#101010] px-5 py-6 text-sm text-white/65">
      <div class="mx-auto flex max-w-7xl flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="flex items-center gap-4">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-[#d4a533]">{{ t('home.docs') }}</a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-[#d4a533]">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const PixelLoadingOverlay = defineAsyncComponent(() => import('@/components/common/PixelLoadingOverlay.vue'))
const showIntroLoading = ref(!sessionStorage.getItem('pixel-home-intro-seen'))
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})
const currentYear = computed(() => new Date().getFullYear())

function initTheme() {
  localStorage.setItem('theme', 'light')
  document.documentElement.classList.remove('dark')
}

onMounted(() => {
  initTheme()
  if (showIntroLoading.value) {
    window.setTimeout(() => {
      showIntroLoading.value = false
      sessionStorage.setItem('pixel-home-intro-seen', '1')
    }, 1450)
  }
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.hero-cta {
  position: relative;
  animation: hero-cta-bob 2.6s steps(4) infinite;
}

.hero-cta::after {
  content: '';
  position: absolute;
  inset: auto 10px -8px;
  height: 6px;
  background: rgba(16, 16, 16, 0.18);
  filter: blur(1px);
  animation: hero-cta-shadow 2.6s steps(4) infinite;
}

.hero-cta:hover {
  animation-play-state: paused;
}

.hero-cta:hover::after {
  animation-play-state: paused;
}

.hero-cta-icon {
  animation: hero-cta-arrow 1.15s steps(3) infinite;
}

.terminal-container {
  position: relative;
}

.terminal-window {
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border: 2px solid #101010;
}

.btn-close {
  background: #a83232;
}

.btn-minimize {
  background: #d4a533;
}

.btn-maximize {
  background: #2d7d46;
}

.terminal-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #a9b4c9;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  letter-spacing: 0.04em;
}

.terminal-body {
  min-height: 280px;
  padding: 24px 20px;
  color: #f8f8f8;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 14px;
  line-height: 1.9;
}

.code-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  opacity: 0;
  animation: line-appear 0.55s steps(5) forwards;
}

.line-1 {
  animation-delay: 0.2s;
}

.line-2 {
  animation-delay: 0.9s;
}

.line-3 {
  animation-delay: 1.6s;
}

.line-4 {
  animation-delay: 2.2s;
}

.code-prompt {
  color: #7df0a8;
  font-weight: 700;
}

.code-cmd {
  color: #7cc2ff;
}

.code-flag {
  color: #f6d06b;
}

.code-url {
  color: #ffd978;
}

.code-comment {
  color: #8f99ad;
}

.code-success {
  border: 2px solid #2d7d46;
  background: rgba(45, 125, 70, 0.18);
  padding: 0 8px;
  color: #7df0a8;
  font-weight: 700;
}

.code-response {
  color: #f8f8f8;
}

.cursor {
  display: inline-block;
  width: 10px;
  height: 18px;
  background: #7df0a8;
  animation: cursor-blink 1s steps(2) infinite;
}

.terminal-metric-label {
  color: #a9b4c9;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.terminal-metric-value {
  margin-top: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 16px;
  font-weight: 800;
}

@keyframes hero-cta-bob {
  0%, 100% {
    transform: translate(0, 0);
  }
  50% {
    transform: translate(-2px, -3px);
  }
}

@keyframes hero-cta-shadow {
  0%, 100% {
    transform: scaleX(1);
    opacity: 0.18;
  }
  50% {
    transform: scaleX(0.92);
    opacity: 0.28;
  }
}

@keyframes hero-cta-arrow {
  0%, 100% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(3px);
  }
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes cursor-blink {
  0%, 49% {
    opacity: 1;
  }
  50%, 100% {
    opacity: 0;
  }
}
</style>
