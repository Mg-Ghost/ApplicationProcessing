<template>
  <div class="page">
    <div class="page-header"><h1>Личный кабинет</h1></div>

    <div class="card" style="max-width:480px;">
      <div class="profile-avatar">{{ initials }}</div>
      <div class="card-title">{{ auth.user?.first_name }} {{ auth.user?.last_name }}</div>
      <p style="font-size:12px;color:var(--text-muted);margin-bottom:20px;">{{ auth.user?.division }}</p>

      <form @submit.prevent="save">
        <div class="form-grid">
          <div class="field"><label>Имя</label><input v-model="f.first_name"></div>
          <div class="field"><label>Фамилия</label><input v-model="f.last_name"></div>
        </div>
        <div class="field">
          <label>Подразделение</label>
          <select v-model="f.division">
            <option v-for="d in divisions" :key="d">{{ d }}</option>
          </select>
        </div>
        <div class="divider"></div>
        <div class="field">
          <label>Новый пароль (оставьте пустым, если не меняете)</label>
          <input type="password" v-model="f.password" minlength="6">
        </div>

        <p v-if="success" style="color:#16a34a;font-size:12px;margin-bottom:10px;">✓ Данные сохранены</p>
        <p v-if="error"   style="color:#dc2626;font-size:12px;margin-bottom:10px;">{{ error }}</p>

        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ saving ? 'Сохранение...' : 'Сохранить изменения' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { profileApi } from '@/api'

const auth  = useAuthStore()
const saving  = ref(false)
const success = ref(false)
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

const f = reactive({ first_name:'', last_name:'', division:'', password:'' })
onMounted(() => {
  f.first_name = auth.user?.first_name || ''
  f.last_name  = auth.user?.last_name  || ''
  f.division   = auth.user?.division   || ''
})

const initials = computed(() =>
  ((auth.user?.first_name?.[0] || '') + (auth.user?.last_name?.[0] || '')).toUpperCase()
)

async function save() {
  error.value = ''
  success.value = false
  saving.value = true
  try {
    await profileApi.update({ ...f })
    await auth.refreshProfile()
    success.value = true
    f.password = ''
  } catch(e) {
    error.value = e.response?.data?.error || 'Ошибка'
  } finally { saving.value = false }
}
</script>

<style scoped>
.profile-avatar {
  width: 64px; height: 64px; border-radius: 50%;
  background: var(--grad-btn); display: flex;
  align-items: center; justify-content: center;
  font-size: 22px; color: white; font-weight: 700;
  margin-bottom: 14px;
}
</style>
