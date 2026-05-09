import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, profileApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const user  = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuth  = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setAuth(t, u) {
    token.value = t
    user.value  = u
    localStorage.setItem('token', t)
    localStorage.setItem('user', JSON.stringify(u))
  }

  function logout() {
    token.value = null
    user.value  = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function register(data) {
    const r = await authApi.register(data)
    setAuth(r.data.token, r.data.user)
  }

  async function login(data) {
    const r = await authApi.login(data)
    setAuth(r.data.token, r.data.user)
  }

  async function adminLogin(data) {
    const r = await authApi.adminLogin(data)
    setAuth(r.data.token, { role: 'admin', login: data.login, ip: r.data.ip })
  }

  async function refreshProfile() {
    const r = await profileApi.get()
    user.value = r.data
    localStorage.setItem('user', JSON.stringify(r.data))
  }

  return { token, user, isAuth, isAdmin, register, login, adminLogin, logout, refreshProfile }
})
