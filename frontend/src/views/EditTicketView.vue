<template>
  <div class="page">
    <div class="page-header">
      <h1>Заявление #{{ id }}</h1>
      <p>
        <span :class="['badge', `badge-${ticket?.status}`]">{{ slabel(ticket?.status) }}</span>
        &nbsp;
        <span :class="['badge', `badge-${ticket?.priority}`]">{{ plabel(ticket?.priority) }}</span>
      </p>
    </div>

    <div v-if="loading" class="card" style="max-width:760px;">Загрузка...</div>

    <div v-else style="max-width:760px;display:flex;flex-direction:column;gap:20px;">

      <!-- Форма редактирования (только если не закрыто/отменено) -->
      <div class="card">
        <div class="card-title" style="font-size:16px;">
          ✏️ Редактировать заявление
          <span v-if="!canEdit" style="font-size:12px;color:var(--text-muted);font-family:'Onest',sans-serif;font-weight:400;margin-left:8px;">
            (недоступно — заявление {{ slabel(ticket.status) }})
          </span>
        </div>

        <form @submit.prevent="save">
          <div class="form-grid">
            <div class="field"><label>Телефон</label>
              <input v-model="f.phone" :disabled="!canEdit"></div>
            <div class="field"><label>Должность</label>
              <input v-model="f.position" :disabled="!canEdit"></div>
            <div class="field"><label>Кабинет</label>
              <input v-model="f.room" :disabled="!canEdit"></div>
            <div class="field"><label>Инвентарный номер</label>
              <input v-model="f.inventory_number" :disabled="!canEdit"></div>
            <div class="field"><label>IP-адрес</label>
              <input v-model="f.ip_address" :disabled="!canEdit"></div>
            <!-- Приоритет задаёт только администратор -->
            <div class="field form-full">
              <label>Описание</label>
              <textarea v-model="f.description" :disabled="!canEdit"></textarea>
            </div>
          </div>

          <p v-if="saveError"   style="color:#dc2626;font-size:12px;margin:8px 0;">{{ saveError }}</p>
          <p v-if="saveSuccess" style="color:#16a34a;font-size:12px;margin:8px 0;">✓ Изменения сохранены</p>

          <div style="display:flex;gap:10px;flex-wrap:wrap;">
            <button v-if="canEdit" type="submit" class="btn btn-primary" :disabled="saving">
              {{ saving ? 'Сохранение...' : 'Сохранить изменения' }}
            </button>
            <router-link to="/dashboard" class="btn btn-ghost">← Назад</router-link>
          </div>
        </form>
      </div>

      <!-- Переписка с IT-отделом -->
      <div class="card" style="padding:0;overflow:hidden;">
        <TicketChat
          :ticket-id="id"
          :messages="ticket.messages || []"
          :status="ticket.status"
          :is-admin="false"
          @sent="reload"
        />
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ticketsApi } from '@/api'
import TicketChat from '@/components/shared/TicketChat.vue'

const route  = useRoute()
const router = useRouter()
const id     = route.params.id

const loading     = ref(true)
const saving      = ref(false)
const saveError   = ref('')
const saveSuccess = ref(false)
const ticket      = ref(null)

const f = reactive({
  phone: '', position: '', room: '', description: '',
  inventory_number: '', ip_address: '', priority: 'medium'
})

const canEdit = computed(() =>
  ticket.value && ticket.value.status !== 'closed' && ticket.value.status !== 'cancelled'
)

async function reload() {
  const r = await ticketsApi.get(id)
  ticket.value = r.data
  Object.assign(f, {
    phone:            r.data.phone,
    position:         r.data.position,
    room:             r.data.room,
    description:      r.data.description,
    inventory_number: r.data.inventory_number,
    ip_address:       r.data.ip_address,
    priority:         r.data.priority,
  })
}

onMounted(async () => {
  await reload()
  loading.value = false
})

async function save() {
  saveError.value   = ''
  saveSuccess.value = false
  saving.value      = true
  try {
    await ticketsApi.update(id, { ...f })
    saveSuccess.value = true
    setTimeout(() => saveSuccess.value = false, 3000)
  } catch(e) {
    saveError.value = e.response?.data?.error || 'Ошибка сохранения'
  } finally { saving.value = false }
}

function slabel(s) {
  return { open:'Открыто', in_progress:'На рассмотрении', closed:'Закрыто', cancelled:'Отменено' }[s] || s
}
function plabel(p) {
  return { high:'Высокий', medium:'Средний', low:'Низкий' }[p] || p
}
</script>
