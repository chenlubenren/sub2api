<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#f8f8f8] p-4 text-[#101010]">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-[linear-gradient(rgba(16,16,16,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(16,16,16,0.045)_1px,transparent_1px)] bg-[size:28px_28px]"
    ></div>

    <router-link
      to="/home"
      class="absolute left-4 top-4 z-20 inline-flex items-center gap-2 border-2 border-[#101010] bg-white px-4 py-2 text-sm font-bold text-[#101010] shadow-[4px_4px_0_#101010] transition-transform hover:-translate-x-0.5 hover:-translate-y-0.5 hover:bg-[#d4a533]"
    >
      <span aria-hidden="true">←</span>
      <span>返回首页</span>
    </router-link>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Card Container -->
      <div class="border-4 border-[#101010] bg-white p-8 shadow-[8px_8px_0_#101010]">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs font-bold text-[#6b6b6b]">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import { normalizeSiteName } from '@/constants/brand'

const appStore = useAppStore()

const siteName = computed(() => normalizeSiteName(appStore.siteName))
const currentYear = computed(() => new Date().getFullYear())
</script>
