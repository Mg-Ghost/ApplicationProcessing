<template>
  <div class="page">
    <div class="page-header">
      <h1>Мои заявления</h1>
      <p>{{ auth.user?.first_name }} {{ auth.user?.last_name }} · {{ auth.user?.division }}</p>
    </div>

    <div v-for="t in escalated" :key="t.id" class="alert-warn">
      ⚠️ Заявление <strong>#{{ t.id }}</strong> не рассмотрено более 5 рабочих дней — приоритет автоматически повышен до <strong>Высокого</strong>
    </div>

    <div class="stats-row">
      <div class="stat-card"><div class="num">{{ tickets.length }}</div><div class="lbl">Всего</div></div>
      <div class="stat-card">
        <div class="num" style="background:var(--grad-warn);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">
          {{ countByStatus('open') + countByStatus('in_progress') }}
        </div>
        <div class="lbl">Активных</div>
      </div>
      <div class="stat-card">
        <div class="num" style="background:var(--grad-success);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">
          {{ countByStatus('closed') }}
        </div>
        <div class="lbl">Закрыто</div>
      </div>
      <div class="stat-card">
        <div class="num" style="background:var(--grad-danger);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">
          {{ countByPriority('high') }}
        </div>
        <div class="lbl">Высокий приоритет</div>
      </div>
    </div>

    <div class="actions-row">
      <router-link to="/tickets/new" class="btn btn-primary">＋ Подать заявление</router-link>
      <button class="btn btn-ghost" :class="{ 'btn-active-filter': filterStatus==='closed' }" @click="toggleFilter('closed')">Закрытые</button>
      <button class="btn btn-ghost" :class="{ 'btn-active-filter': filterStatus==='' }"      @click="toggleFilter('')">Все</button>
    </div>

    <div v-if="loading" class="loading-msg">Загрузка...</div>

    <div v-else-if="filtered.length === 0" class="empty-msg">
      Заявлений нет. <router-link to="/tickets/new">Подайте первое заявление</router-link>
    </div>

    <div v-else class="table-wrap">
      <table>
        <thead><tr>
          <th>#</th><th>Описание</th><th>Подразделение</th><th>Дата</th>
          <th>Приоритет</th><th>Статус</th><th>Действия</th>
        </tr></thead>
        <tbody>
          <tr v-for="t in filtered" :key="t.id">
            <td><strong>#{{ t.id }}</strong></td>
            <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ t.description }}</td>
            <td>{{ t.division }}</td>
            <td>{{ formatDate(t.created_at) }}</td>
            <td><span :class="['badge', `badge-${t.priority}`]">{{ priorityLabel(t.priority) }}</span></td>
            <td><span :class="['badge', `badge-${t.status}`]">{{ statusLabel(t.status) }}</span></td>
            <td>
              <div style="display:flex;gap:6px;flex-wrap:wrap;align-items:center;">

                <!-- Просмотр с бейджем непрочитанных -->
                <span class="btn-with-badge" @click="openTicket(t.id)">
                  <span class="btn btn-ghost btn-sm">👁 Просмотр</span>
                  <span v-if="unread[t.id]" class="unread-badge">{{ unread[t.id] }}</span>
                </span>

                <!-- Редактировать + переписка -->
                <router-link
                  v-if="t.status === 'open' || t.status === 'in_progress'"
                  :to="`/tickets/${t.id}/edit`"
                  class="btn btn-primary btn-sm">✏️ Редактировать</router-link>

                <!-- Отменить -->
                <button
                  v-if="t.status === 'open'"
                  class="btn btn-warn btn-sm"
                  @click="cancelTicket(t.id)">Отменить</button>

                <!-- Закрыть -->
                <button
                  v-if="t.status === 'open' || t.status === 'in_progress'"
                  class="btn btn-ghost btn-sm"
                  @click="closeTicket(t.id)">✓ Закрыть</button>

              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ticketsApi } from '@/api'

const auth         = useAuthStore()
const router       = useRouter()
const tickets      = ref([])
const unread       = ref({})
const loading      = ref(true)
const filterStatus = ref('')

const filtered = computed(() => {
  if (!filterStatus.value) return tickets.value
  return tickets.value.filter(t => t.status === filterStatus.value)
})

const escalated = computed(() =>
  tickets.value.filter(t => t.auto_escalated && t.status !== 'closed' && t.status !== 'cancelled')
)

function countByStatus(s)  { return tickets.value.filter(t => t.status === s).length }
function countByPriority(p){ return tickets.value.filter(t => t.priority === p).length }
function toggleFilter(s)   { filterStatus.value = filterStatus.value === s ? '' : s }

// Открыть заявление — убрать бейдж сразу, перейти на страницу
function openTicket(id) {
  const updated = { ...unread.value }
  delete updated[id]
  unread.value = updated
  router.push(`/tickets/${id}`)
}

async function load() {
  loading.value = true
  try {
    const [r, u] = await Promise.all([
      ticketsApi.list(),
      ticketsApi.unreadCounts()
    ])
    tickets.value = r.data || []
    unread.value  = u.data || {}
  } catch(e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function cancelTicket(id) {
  if (!confirm('Отменить заявление?')) return
  await ticketsApi.cancel(id)
  await load()
}

async function closeTicket(id) {
  if (!confirm('Закрыть заявление (проблема решена)?')) return
  await ticketsApi.close(id)
  await load()
}

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function priorityLabel(p) { return { high:'Высокий ⚡', medium:'Средний', low:'Низкий' }[p] || p }
function statusLabel(s)   { return { open:'Открыто', in_progress:'На рассмотрении', closed:'Закрыто', cancelled:'Отменено' }[s] || s }

let pollInterval = null

onMounted(async () => {
  await load()
  // Polling: проверяем новые уведомления каждые 15 секунд
  pollInterval = setInterval(async () => {
    try {
      const u = await ticketsApi.unreadCounts()
      // Обновляем только те заявки у которых изменилось число — не трогаем уже убранные
      const fresh = u.data || {}
      const merged = { ...unread.value }
      Object.keys(fresh).forEach(id => { merged[id] = fresh[id] })
      unread.value = merged
    } catch(e) { /* тихо игнорируем */ }
  }, 15000)
})

onUnmounted(() => {
  clearInterval(pollInterval)
})
</script>

<style scoped>
.loading-msg, .empty-msg { text-align:center; color:var(--text-muted); padding:40px; font-size:14px; }
.empty-msg a { color: var(--accent); }
.btn-active-filter { background:var(--surface2); color:var(--text); border-color:var(--accent); }

.btn-with-badge {
  position: relative;
  display: inline-flex;
  cursor: pointer;
}
.unread-badge {
  position: absolute; top: -7px; right: -7px;
  background: #ef4444; color: white;
  font-size: 10px; font-weight: 700;
  min-width: 18px; height: 18px;
  border-radius: 9px; display: flex;
  align-items: center; justify-content: center;
  padding: 0 4px; border: 2px solid white;
  pointer-events: none; z-index: 1;
}
</style>