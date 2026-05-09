<template>
  <div class="page">
    <div class="page-header">
      <h1>Изменить заявление #{{ id }}</h1>
    </div>

    <div v-if="loading" class="card" style="max-width:740px;">Загрузка...</div>

    <div v-else class="card" style="max-width:740px;">
      <form @submit.prevent="submit">
        <div class="form-grid">
          <div class="field"><label>Телефон</label><input v-model="f.phone"></div>
          <div class="field"><label>Должность</label><input v-model="f.position"></div>
          <div class="field"><label>Кабинет</label><input v-model="f.room"></div>
          <div class="field"><label>Инвентарный номер</label><input v-model="f.inventory_number"></div>
          <div class="field"><label>IP-адрес</label><input v-model="f.ip_address"></div>
          <div class="field">
            <label>Приоритет</label>
            <select v-model="f.priority">
              <option value="low">Низкий</option>
              <option value="medium">Средний</option>
              <option value="high">Высокий</option>
            </select>
          </div>
          <div class="field form-full">
            <label>Описание</label>
            <textarea v-model="f.description"></textarea>
          </div>
        </div>

        <p v-if="error" style="color:#dc2626;font-size:12px;margin-bottom:10px;">{{ error }}</p>

        <div style="display:flex;gap:10px;margin-top:8px;">
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? 'Сохранение...' : 'Сохранить' }}</button>
          <router-link to="/dashboard" class="btn btn-ghost">Отмена</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ticketsApi } from '@/api'

const route  = useRoute()
const router = useRouter()
const id     = route.params.id

const loading = ref(true)
const saving  = ref(false)
const error   = ref('')
const f = reactive({ phone:'', position:'', room:'', description:'', inventory_number:'', ip_address:'', priority:'medium' })

onMounted(async () => {
  const r = await ticketsApi.get(id)
  Object.assign(f, r.data)
  loading.value = false
})

async function submit() {
  error.value = ''
  saving.value = true
  try {
    await ticketsApi.update(id, { ...f })
    router.push('/dashboard')
  } catch(e) {
    error.value = e.response?.data?.error || 'Ошибка'
  } finally { saving.value = false }
}
</script>
