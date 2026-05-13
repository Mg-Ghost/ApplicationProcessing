<template>
  <div class="page">
    <div class="page-header">
      <h1>Новое заявление</h1>
      <p>Заполните все обязательные поля</p>
    </div>

    <div class="card" style="max-width:740px;">
      <form @submit.prevent="submit">

        <div class="section-label">Контактные данные</div>
        <div class="form-grid">
          <div class="field"><label>Имя *</label><input v-model="f.first_name" required></div>
          <div class="field"><label>Фамилия *</label><input v-model="f.last_name" required></div>
          <div class="field"><label>Номер телефона *</label><input v-model="f.phone" type="tel" placeholder="+7 999 000-00-00" required></div>
          <div class="field"><label>Должность *</label><input v-model="f.position" required></div>
        </div>

        <div class="divider"></div>
        <div class="section-label">Место обращения</div>
        <div class="form-grid">
          <div class="field"><label>Кабинет *</label><input v-model="f.room" placeholder="Кабинет 214" required></div>
          <div class="field">
            <label>Подразделение *</label>
            <select v-model="f.division" required>
              <option value="">Выберите...</option>
              <option v-for="d in divisions" :key="d">{{ d }}</option>
            </select>
          </div>
        </div>

        <div class="divider"></div>
        <div class="section-label">Описание проблемы</div>
        <div class="form-grid">
          <div class="field form-full">
            <label>Категория *</label>
            <select v-model="f.category" required>
              <option value="">Выберите категорию...</option>
              <option>РМИС</option>
              <option>ПАРУС</option>
              <option>РУЧНОЙ СКАНЕР</option>
              <option>ПРИНТЕР</option>
              <option>КОМП</option>
              <option>КОНСУЛЬТАНТ</option>
              <option>ЭЦП</option>
              <option>Другое</option>
            </select>
          </div>
          <div class="field form-full">
            <label>Описание * (укажите фирму компьютера)</label>
            <textarea v-model="f.description" placeholder="Опишите проблему подробно. Укажите марку и модель оборудования (Dell Optiplex, HP EliteDesk и т.д.)" minlength="10" required></textarea>
          </div>
          <div class="field">
            <label>Инвентарный номер</label>
            <input v-model="f.inventory_number" placeholder="ИНВ-00412">
          </div>
          <div class="field">
            <label>IP-адрес (при наличии)</label>
            <input v-model="f.ip_address" placeholder="192.168.1.45">
          </div>
          <!-- Приоритет задаёт администратор после рассмотрения заявки -->
        </div>

        <p v-if="error" style="color:#dc2626;font-size:12px;margin-bottom:10px;">{{ error }}</p>

        <div style="display:flex;gap:10px;margin-top:8px;flex-wrap:wrap;">
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? 'Отправка...' : '📨 Подать заявление' }}
          </button>
          <router-link to="/dashboard" class="btn btn-ghost">Отмена</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ticketsApi } from '@/api'

const router = useRouter()
const auth   = useAuthStore()

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

const f = reactive({
  first_name: auth.user?.first_name || '',
  last_name:  auth.user?.last_name  || '',
  phone:      '',
  position:   '',
  room:       '',
  division:   auth.user?.division   || '',
  category:   '',
  description:'',
  inventory_number: '',
  ip_address: '',
})

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await ticketsApi.create({ ...f })
    router.push('/dashboard')
  } catch(e) {
    error.value = e.response?.data?.error || 'Ошибка при отправке'
  } finally { loading.value = false }
}
</script>
