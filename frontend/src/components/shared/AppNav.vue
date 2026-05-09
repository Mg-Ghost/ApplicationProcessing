<template>
  <nav class="app-nav">
    <div class="nav-brand">
      <span class="nav-logo">МедДок</span>
      <span v-if="auth.isAdmin" class="admin-chip">ADMIN</span>
    </div>

    <div class="nav-links">
      <template v-if="!auth.isAdmin">
        <router-link to="/dashboard" class="nav-link">Мои заявления</router-link>
        <router-link to="/tickets/new" class="nav-link">Подать заявление</router-link>
        <router-link to="/profile" class="nav-link">Профиль</router-link>
      </template>
      <template v-else>
        <router-link to="/admin" class="nav-link">Панель управления</router-link>
      </template>
    </div>

    <button class="btn btn-ghost btn-sm" @click="handleLogout">Выйти</button>
  </nav>
</template>

<script setup>
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-nav {
  display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
  padding: 14px 32px; background: var(--surface);
  border-bottom: 1px solid var(--border);
  box-shadow: 0 1px 6px rgba(0,0,0,.04);
  position: sticky; top: 0; z-index: 100;
}
.nav-brand { display: flex; align-items: center; gap: 8px; margin-right: 16px; }
.nav-logo { font-family: 'Playfair Display', serif; font-size: 17px; color: var(--accent); font-weight: 600; }
.admin-chip {
  background: var(--grad-dark); color: #a0a8ff;
  font-size: 9px; padding: 2px 7px; border-radius: 20px;
  font-weight: 700; letter-spacing: .8px;
}
.nav-links { display: flex; gap: 4px; flex: 1; }
.nav-link {
  padding: 7px 14px; border-radius: 8px; font-size: 13px; font-weight: 500;
  color: var(--text-muted); text-decoration: none; transition: all .2s;
}
.nav-link:hover { background: var(--surface2); color: var(--text); }
.nav-link.router-link-active { background: var(--grad-btn); color: white; box-shadow: 0 2px 8px rgba(79,124,255,.25); }
</style>
