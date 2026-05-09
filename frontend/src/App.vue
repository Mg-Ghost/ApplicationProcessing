<template>
  <AppNav v-if="auth.isAuth" />
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup>
import { useAuthStore } from '@/stores/auth'
import AppNav from '@/components/shared/AppNav.vue'
const auth = useAuthStore()
</script>
