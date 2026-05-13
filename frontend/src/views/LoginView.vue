<template>
  <div class="login-page">
    <div class="login-hero">
      <div class="hero-text">
        <h1>МедДок</h1>
        <p>Система управления заявками<br>IT-службы медицинского учреждения</p>
      </div>
    </div>

    <div class="login-panel">
      <div class="tab-row">
        <button :class="['tab', {active: tab==='register'}]" @click="tab='register'">Регистрация</button>
        <button :class="['tab', {active: tab==='login'}]"    @click="tab='login'">Вход</button>
        <button :class="['tab', {active: tab==='admin'}]"    @click="tab='admin'">Администратор</button>
      </div>

      <!-- REGISTER -->
      <form v-if="tab==='register'" @submit.prevent="doRegister" class="auth-form">
        <h2 class="card-title">Создать аккаунт</h2>
        <div class="form-grid">
          <div class="field"><label>Имя</label><input v-model="reg.first_name" placeholder="Иван" required></div>
          <div class="field"><label>Фамилия</label><input v-model="reg.last_name" placeholder="Петров" required></div>
        </div>
        <div class="field">
          <label>Подразделение</label>
          <select v-model="reg.division" required>
            <option value="">Выберите подразделение</option>
            <option v-for="d in divisions" :key="d">{{ d }}</option>
          </select>
        </div>
        <div class="form-grid">
          <div class="field"><label>Пароль</label><input type="password" v-model="reg.password" minlength="6" required></div>
          <div class="field"><label>Повторите пароль</label><input type="password" v-model="reg.confirm" required></div>
        </div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn btn-success btn-full" :disabled="loading">
          {{ loading ? 'Загрузка...' : 'Зарегистрироваться' }}
        </button>
      </form>

      <!-- LOGIN -->
      <form v-else-if="tab==='login'" @submit.prevent="doLogin" class="auth-form">
        <h2 class="card-title">Добро пожаловать</h2>
        <div class="form-grid">
          <div class="field"><label>Имя</label><input v-model="log.first_name" placeholder="Иван" required></div>
          <div class="field"><label>Фамилия</label><input v-model="log.last_name" placeholder="Петров" required></div>
        </div>
        <div class="field"><label>Пароль</label><input type="password" v-model="log.password" required></div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn btn-primary btn-full" :disabled="loading">
          {{ loading ? 'Загрузка...' : 'Войти' }}
        </button>
      </form>

      <!-- ADMIN -->
      <form v-else @submit.prevent="doAdminLogin" class="auth-form">
        <h2 class="card-title">Вход администратора</h2>
        <div class="field"><label>Логин (латиница)</label><input v-model="adm.login" placeholder="admin_user" pattern="[a-zA-Z0-9_]+" required></div>
        <div class="field"><label>Пароль (мин. 10 символов)</label><input type="password" v-model="adm.password" minlength="10" required></div>
        <div class="field"><label>Секретный ключ (мин. 8 символов)</label><input type="password" v-model="adm.secret_key" minlength="8" required></div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn btn-dark btn-full" :disabled="loading">
          {{ loading ? 'Проверка...' : 'Войти как администратор' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth   = useAuthStore()

const tab     = ref('login')
const loading = ref(false)
const error   = ref('')

const divisions = [
  'ДС при ПВ№3', 'ДС при ПВ№3 неврология', 'Поликлиника взрослых №3',
  'Отд. поликлиника взрослых №3', 'Детская поликлиника №2', 'Детская поликлиника №3',
  'Детская поликлиника №4', 'Пульмоцентр', 'Центр здоровья', 'ПСЦ Неврология',
  'Колопроктология', 'Хирургия №1', 'Хирургия №2', 'ДИБО', 'Урология',
  'Гинекология', 'Гастро отделение', 'Неврология', 'Приемное отделение',
  'Терапия', 'Реанимация', 'Пульмо отделение', 'ЛОР', 'Кабинет трансфузии',
  'КДЛ', 'Операционный блок', 'Рентген отделение', 'Физиотерапия', 'ЦСО',
  'Эндоскопическое отделение', 'ФД и УЗИ', 'Отделение реабилитации',
  'Аптека', 'Администрация', 'Баклаборатория', 'Школа сахарного диабета',
]

const reg = reactive({ first_name:'', last_name:'', division:'', password:'', confirm:'' })
const log = reactive({ first_name:'', last_name:'', password:'' })
const adm = reactive({ login:'', password:'', secret_key:'' })

async function doRegister() {
  error.value = ''
  if (reg.password !== reg.confirm) { error.value = 'Пароли не совпадают'; return }
  loading.value = true
  try {
    await auth.register({ first_name: reg.first_name, last_name: reg.last_name, division: reg.division, password: reg.password })
    router.push('/dashboard')
  } catch(e) {
    error.value = e.response?.data?.error || 'Ошибка регистрации'
  } finally { loading.value = false }
}

async function doLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login({ first_name: log.first_name, last_name: log.last_name, password: log.password })
    router.push('/dashboard')
  } catch(e) {
    error.value = e.response?.data?.error || 'Неверные данные'
  } finally { loading.value = false }
}

async function doAdminLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.adminLogin({ login: adm.login, password: adm.password, secret_key: adm.secret_key })
    router.push('/admin')
  } catch(e) {
    error.value = e.response?.data?.error || 'Неверные данные'
  } finally { loading.value = false }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh; display: flex;
}
.login-hero {
  flex: 1; display: flex; align-items: center; justify-content: center;
  background: var(--grad-dark); padding: 48px;
}
.hero-text h1 {
  font-family: 'Playfair Display', serif;
  font-size: 48px; color: white; margin-bottom: 16px;
}
.hero-text p { color: rgba(255,255,255,.55); font-size: 16px; line-height: 1.7; }

.login-panel {
  width: 480px; background: var(--surface);
  display: flex; flex-direction: column; justify-content: center;
  padding: 48px 40px; box-shadow: -8px 0 40px rgba(0,0,0,.06);
}
.tab-row { display: flex; gap: 4px; margin-bottom: 28px; border-bottom: 2px solid var(--border); padding-bottom: 14px; }
.tab {
  padding: 7px 14px; border-radius: 8px; border: none; cursor: pointer;
  font-family: 'Onest', sans-serif; font-size: 12px; font-weight: 600;
  color: var(--text-muted); background: transparent; transition: all .2s;
}
.tab.active { background: var(--grad-btn); color: white; }
.tab:hover:not(.active) { background: var(--surface2); color: var(--text); }

.auth-form { display: flex; flex-direction: column; gap: 0; }
.error-msg { color: #dc2626; font-size: 12px; margin-bottom: 8px; }

@media (max-width: 768px) {
  .login-page { flex-direction: column; }
  .login-hero { padding: 32px 24px; min-height: 180px; }
  .login-panel { width: 100%; padding: 28px 20px; }
  .hero-text h1 { font-size: 32px; }
}
</style>
