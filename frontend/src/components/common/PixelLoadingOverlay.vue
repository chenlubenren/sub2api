<template>
  <Transition name="pixel-loading">
    <div
      v-if="visible"
      class="pointer-events-none fixed inset-0 z-[90] flex items-center justify-center overflow-hidden bg-[#f8f8f8]"
      role="status"
      aria-live="polite"
    >
      <div class="absolute inset-0 bg-[linear-gradient(rgba(16,16,16,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(16,16,16,0.05)_1px,transparent_1px)] bg-[size:24px_24px]"></div>

      <svg class="absolute inset-0 h-full w-full opacity-30" aria-hidden="true">
        <line
          v-for="edge in edges"
          :key="edge.key"
          :x1="`${edge.from.x}%`"
          :y1="`${edge.from.y}%`"
          :x2="`${edge.to.x}%`"
          :y2="`${edge.to.y}%`"
          :stroke="edge.color"
          stroke-width="2"
          class="pixel-loading-edge"
          :style="{ animationDelay: `${edge.delay}s` }"
        />
      </svg>

      <span
        v-for="node in nodes"
        :key="node.id"
        class="pixel-loading-node absolute border-2 border-[#101010]"
        :style="{
          left: `${node.x}%`,
          top: `${node.y}%`,
          width: `${node.size}px`,
          height: `${node.size}px`,
          backgroundColor: node.color,
          animationDelay: `${node.delay}s`
        }"
      ></span>

      <span
        v-for="packet in packets"
        :key="packet.id"
        class="pixel-loading-packet absolute h-2 w-2 bg-[#2d7d46]"
        :style="{
          left: `${packet.x}%`,
          top: `${packet.y}%`,
          animationDelay: `${packet.delay}s`
        }"
      ></span>

      <div class="relative z-10 w-[min(420px,calc(100vw-40px))] border-4 border-[#101010] bg-white p-7 text-center shadow-[10px_10px_0_#101010]">
        <div class="mx-auto mb-5 flex h-20 w-20 items-center justify-center border-4 border-[#101010] bg-[#3a5ba0] shadow-[5px_5px_0_#101010] pixel-loading-core">
          <span class="grid grid-cols-3 gap-1">
            <span v-for="dot in 9" :key="dot" class="h-2.5 w-2.5 bg-[#f8f8f8]"></span>
          </span>
        </div>

        <h2 class="text-2xl font-black tracking-normal text-[#101010]">助研算力供应中心</h2>
        <p class="mt-2 text-sm font-semibold uppercase tracking-normal text-[#3a5ba0]">POWER ROUTING...</p>

        <div class="mx-auto mt-5 h-4 w-56 overflow-hidden border-2 border-[#101010] bg-[#f8f8f8]">
          <div class="pixel-loading-bar h-full bg-[#2d7d46]"></div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineProps<{
  visible: boolean
}>()

const palette = ['#3A5BA0', '#2D7D46', '#A83232', '#D4A533', '#6B6B6B']

const nodes = Array.from({ length: 18 }, (_, index) => ({
  id: index,
  x: 8 + ((index * 23) % 84),
  y: 9 + ((index * 37) % 78),
  size: 8 + (index % 4) * 3,
  delay: (index % 7) * 0.18,
  color: palette[index % palette.length]
}))

const edges = computed(() =>
  nodes.slice(0, -1).map((node, index) => ({
    key: `${node.id}-${nodes[index + 1].id}`,
    from: node,
    to: nodes[index + 1],
    color: index % 2 === 0 ? '#3A5BA0' : '#2D7D46',
    delay: (index % 6) * 0.22
  }))
)

const packets = Array.from({ length: 12 }, (_, index) => ({
  id: index,
  x: 10 + ((index * 29) % 80),
  y: 12 + ((index * 41) % 74),
  delay: index * 0.24
}))
</script>

<style scoped>
.pixel-loading-enter-active,
.pixel-loading-leave-active {
  transition: opacity 0.35s ease;
}

.pixel-loading-enter-from,
.pixel-loading-leave-to {
  opacity: 0;
}

.pixel-loading-node {
  transform: translate(-50%, -50%);
  animation: pixel-node-pulse 2.4s steps(4) infinite;
}

.pixel-loading-edge {
  animation: pixel-edge-pulse 2.2s steps(4) infinite;
}

.pixel-loading-packet {
  animation: pixel-packet-float 4.5s steps(8) infinite;
}

.pixel-loading-core {
  animation: pixel-core-bob 2s steps(4) infinite;
}

.pixel-loading-bar {
  animation: pixel-bar-load 1.15s steps(8) forwards;
}

@keyframes pixel-node-pulse {
  0%, 100% { opacity: 0.4; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 1; transform: translate(-50%, -50%) scale(1.35); }
}

@keyframes pixel-edge-pulse {
  0%, 100% { opacity: 0.18; }
  50% { opacity: 0.72; }
}

@keyframes pixel-packet-float {
  0% { opacity: 0; transform: translate(0, 0); }
  20% { opacity: 0.75; }
  80% { opacity: 0.45; }
  100% { opacity: 0; transform: translate(42px, -28px); }
}

@keyframes pixel-core-bob {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

@keyframes pixel-bar-load {
  from { width: 0; }
  to { width: 100%; }
}
</style>
