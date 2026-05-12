<template>
  <div class="page">
    <!-- Header -->
    <div class="admin-header">
      <div>
        <h1>Панель администратора</h1>
        <p>Управление заявками · Все подразделения</p>
      </div>
      <div class="ip-info">
        <div class="ip-tag">🌐 Последний вход: {{ lastIP }}</div>
        <div class="ip-time">{{ loginTime }}</div>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card clickable" @click="setFilter('')">
        <div class="num">{{ tickets.length }}</div>
        <div class="lbl">Всего заявлений</div>
      </div>
      <div class="stat-card clickable" @click="setFilter('open')" :class="{active: filter.status==='open'}">
        <div class="num" style="background:var(--grad-warn);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">{{ count('status','open') }}</div>
        <div class="lbl">Открытые</div>
      </div>
      <div class="stat-card clickable" @click="setFilter('in_progress')" :class="{active: filter.status==='in_progress'}">
        <div class="num" style="background:linear-gradient(135deg,#7c5cfc,#5cbbfc);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">{{ count('status','in_progress') }}</div>
        <div class="lbl">На рассмотрении</div>
      </div>
      <div class="stat-card clickable" @click="setFilter('high_priority')" :class="{active: filter.priority==='high'}">
        <div class="num" style="background:var(--grad-danger);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">{{ count('priority','high') }}</div>
        <div class="lbl">Высокий приоритет</div>
      </div>
      <div class="stat-card clickable" @click="setFilter('closed')" :class="{active: filter.status==='closed'}">
        <div class="num" style="background:var(--grad-success);-webkit-background-clip:text;-webkit-text-fill-color:transparent;">{{ count('status','closed') }}</div>
        <div class="lbl">Закрытые</div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <select v-model="filter.division"   @change="load"><option value="">Все подразделения</option><option v-for="d in divisions" :key="d">{{ d }}</option></select>
      <select v-model="filter.priority"   @change="load"><option value="">Все приоритеты</option><option value="high">Высокий</option><option value="medium">Средний</option><option value="low">Низкий</option></select>
      <select v-model="filter.status"     @change="load"><option value="">Все статусы</option><option value="open">Открыто</option><option value="in_progress">На рассмотрении</option><option value="closed">Закрыто</option><option value="cancelled">Отменено</option></select>
      <input type="date" v-model="filter.date_from" @change="load">
      <input type="date" v-model="filter.date_to"   @change="load">
      <select v-model="filter.sort_by"    @change="load"><option value="">Сортировка</option><option value="created_at">Дата</option><option value="priority">Приоритет</option><option value="division">Подразделение</option></select>
      <select v-model="filter.sort_order" @change="load"><option value="DESC">↓ Убыв.</option><option value="ASC">↑ Возр.</option></select>
      <button class="btn btn-ghost btn-sm" @click="resetFilter">Сбросить</button>
    </div>

    <!-- Export -->
    <div class="export-row">
      <span style="font-size:12px;color:var(--text-muted);align-self:center;">Экспорт:</span>
      <button class="btn btn-ghost btn-sm" @click="doExportXLSX">⬇ XLSX</button>
    </div>

    <!-- Table -->
    <div class="table-wrap" style="margin-bottom:24px;">
      <table>
        <thead><tr>
          <th>#</th><th>ФИО</th><th>Категория</th><th>Описание</th><th>Подразделение</th>
          <th>Дата</th><th>Приоритет</th><th>Статус</th><th>Закрыл</th><th>Действия</th>
        </tr></thead>
        <tbody>
          <tr v-for="t in tickets" :key="t.id">
            <td><router-link :to="`/tickets/${t.id}`" style="font-weight:700;color:var(--accent);text-decoration:none;">#{{ t.id }}</router-link></td>
            <td>{{ t.first_name }} {{ t.last_name }}</td>
            <td><span style="display:inline-block;padding:2px 7px;background:var(--surface2);border:1px solid var(--border);border-radius:6px;font-size:11px;font-weight:600;color:var(--text-muted);">{{ t.category || '—' }}</span></td>
            <td class="desc-cell" :title="t.description">{{ t.description }}</td>
            <td>{{ t.division }}</td>
            <td>{{ fmt(t.created_at) }}</td>
            <td>
              <span :class="['badge',`badge-${t.priority}`]">{{ plabel(t.priority) }}</span>
              <span v-if="t.auto_escalated" title="Автоэскалация" style="margin-left:4px;">⚡</span>
            </td>
            <td><span :class="['badge',`badge-${t.status}`]">{{ slabel(t.status) }}</span></td>
            <td style="font-size:12px;">
              <span v-if="t.closed_by_admin">
                <span :style="t.closed_by_role==='admin' ? 'color:var(--accent);font-weight:600;' : 'color:#16a34a;font-weight:600;'">
                  {{ t.closed_by_role === 'admin' ? '🔧' : '👤' }}
                  <strong>{{ t.closed_by_admin }}</strong>
                </span>
                <div style="font-size:10px;color:var(--text-muted);margin-top:1px;">
                  {{ t.closed_by_role === 'admin' ? 'администратор' : 'пользователь' }}
                </div>
              </span>
              <span v-else style="color:var(--text-muted);">—</span>
            </td>
            <td>
              <div style="display:flex;gap:4px;flex-wrap:wrap;">
                <router-link :to="`/tickets/${t.id}`" class="btn btn-ghost btn-sm" title="Просмотр">👁</router-link>
                  <span class="btn-with-badge">
                  <button class="btn btn-primary btn-sm" style="padding: 3.5px 7.5px;" @click="openChat(t)" title="Открыть переписку">💬 Ответить</button>
                  <span v-if="unread[t.id]" class="unread-badge">{{ unread[t.id] }}</span>
                </span>
                <button v-if="t.status !== 'closed'" class="btn btn-ghost btn-sm" @click="closeT(t.id)" title="Закрыть">✓</button>
                <button class="btn btn-danger btn-sm" @click="deleteT(t.id)" title="Удалить">🗑</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && tickets.length===0">
            <td colspan="8" style="text-align:center;color:var(--text-muted);padding:32px;">Заявлений нет</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Chat modal -->
    <div v-if="chatModal" class="modal-overlay" @click.self="chatModal=false; stopChatPolling()">
      <div class="card modal-box">
        <div class="modal-top">
          <div>
            <h3 class="card-title" style="margin:0;font-size:17px;">Заявление #{{ activeTicket?.id }}</h3>
            <div class="ticket-info-row">
              <span>👤 {{ activeTicket?.first_name }} {{ activeTicket?.last_name }}</span>
              <span>🏥 {{ activeTicket?.division }}</span>
              <span>🚪 {{ activeTicket?.room }}</span>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" @click="chatModal=false; stopChatPolling()">✕</button>
        </div>

        <div class="ticket-desc">{{ activeTicket?.description }}</div>

        <!-- Смена приоритета администратором -->
        <div style="display:flex;align-items:center;gap:10px;margin-bottom:12px;flex-wrap:wrap;">
          <span style="font-size:12px;color:var(--text-muted);font-weight:600;">ПРИОРИТЕТ:</span>
          <select v-model="editPriority" @change="changePriority" style="padding:5px 10px;border-radius:8px;border:1.5px solid var(--border);font-family:'Onest',sans-serif;font-size:12px;outline:none;">
            <option value="low">Низкий</option>
            <option value="medium">Средний</option>
            <option value="high">Высокий</option>
          </select>
          <span v-if="prioritySaved" style="font-size:12px;color:#16a34a;">✓ Сохранено</span>
        </div>

        <div style="margin-bottom:8px;display:flex;gap:8px;flex-wrap:wrap;">
          <span :class="['badge', `badge-${activeTicket?.priority}`]">{{ plabel(activeTicket?.priority) }}</span>
          <span :class="['badge', `badge-${activeTicket?.status}`]">{{ slabel(activeTicket?.status) }}</span>
          <span v-if="activeTicket?.inventory_number" style="font-size:12px;color:var(--text-muted);">
            🏷 {{ activeTicket.inventory_number }}
          </span>
          <span v-if="activeTicket?.ip_address" style="font-size:12px;color:var(--text-muted);">
            🌐 {{ activeTicket.ip_address }}
          </span>
        </div>

        <TicketChat
          v-if="activeTicket"
          :ticket-id="activeTicket.id"
          :messages="activeTicket.messages || []"
          :status="activeTicket.status"
          :is-admin="true"
          @sent="onAdminSent"
        />
      </div>
    </div>

    <!-- IP logs -->
    <div class="card">
      <div class="card-title" style="font-size:15px;">🌐 Журнал входов администратора</div>
      <table>
        <thead><tr><th>#</th><th>Логин</th><th>IP-адрес</th><th>Дата/время</th></tr></thead>
        <tbody>
          <tr v-for="l in ipLogs" :key="l.id">
            <td>{{ l.id }}</td>
            <td>{{ l.login }}</td>
            <td><code style="font-size:12px;">{{ l.ip_address }}</code></td>
            <td>{{ fmt(l.created_at) }} {{ fmtTime(l.created_at) }}</td>
          </tr>
          <tr v-if="!ipLogs.length"><td colspan="4" style="color:var(--text-muted);text-align:center;padding:20px;">Нет записей</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { adminApi } from '@/api'
import TicketChat from '@/components/shared/TicketChat.vue'

const tickets   = ref([])
const unread    = ref({})
const ipLogs    = ref([])
const loading   = ref(true)
const chatModal    = ref(false)
const editPriority  = ref('medium')
const prioritySaved = ref(false)
const activeTicket = ref(null)
const lastIP    = ref('—')
const loginTime = ref('—')

const filter = reactive({
  division:'', priority:'', status:'',
  date_from:'', date_to:'', sort_by:'', sort_order:'DESC'
})

const divisions = [
  'ГКБ №1 — Хирургия', 'ГКБ №1 — Терапия', 'ГКБ №1 — Кардиология',
  'ГКБ №2 — Хирургия', 'ГКБ №2 — Неврология', 'ГКБ №3 — Реанимация',
  'ГКБ №3 — Педиатрия', 'ГКБ №4 — Онкология',
]

async function load() {
  loading.value = true
  const params = {}
  Object.entries(filter).forEach(([k,v]) => { if(v) params[k] = v })
  const [r, u] = await Promise.all([adminApi.listTickets(params), adminApi.unreadCounts()])
  tickets.value = r.data || []
  unread.value  = u.data || {}
  loading.value = false
}

async function loadIPLogs() {
  const r = await adminApi.ipLogs()
  ipLogs.value = r.data || []
  if (ipLogs.value.length) {
    lastIP.value   = ipLogs.value[0].ip_address
    loginTime.value = fmt(ipLogs.value[0].created_at) + ' ' + fmtTime(ipLogs.value[0].created_at)
  }
}

function setFilter(val) {
  // Сбрасываем все фильтры, ставим нужный
  Object.assign(filter, { division:'', priority:'', status:'', date_from:'', date_to:'', sort_by:'', sort_order:'DESC' })
  if (val === 'high_priority') {
    filter.priority = 'high'
  } else if (val) {
    filter.status = val
  }
  load()
}

function resetFilter() {
  Object.assign(filter, { division:'', priority:'', status:'', date_from:'', date_to:'', sort_by:'', sort_order:'DESC' })
  load()
}

async function changePriority() {
  if (!activeTicket.value) return
  await adminApi.setPriority(activeTicket.value.id, editPriority.value)
  activeTicket.value.priority = editPriority.value
  // Обновляем в таблице тоже
  const found = tickets.value.find(t => t.id === activeTicket.value.id)
  if (found) found.priority = editPriority.value
  prioritySaved.value = true
  setTimeout(() => prioritySaved.value = false, 2000)
}

async function openChat(ticket) {
  try {
    const r = await adminApi.getTicket(ticket.id)
    activeTicket.value = r.data
    editPriority.value  = r.data.priority || 'medium'
    prioritySaved.value = false
    chatModal.value    = true
    startChatPolling() // начинаем опрос новых сообщений
    // Сразу убираем бейдж для этой заявки — мы её прочитали
    const updated = { ...unread.value }
    delete updated[ticket.id]
    unread.value = updated
  } catch(e) {
    alert('Ошибка загрузки заявки: ' + (e.response?.data?.error || e.message))
  }
}

async function onAdminSent() {
  if (activeTicket.value) {
    const r = await adminApi.getTicket(activeTicket.value.id)
    activeTicket.value = r.data
    // Убираем бейдж только для текущей заявки, не трогаем остальные
    const updated = { ...unread.value }
    delete updated[activeTicket.value.id]
    unread.value = updated
  }
  // Обновляем список заявок без перезаписи unread
  const params = {}
  Object.entries(filter).forEach(([k,v]) => { if(v) params[k] = v })
  const r = await adminApi.listTickets(params)
  tickets.value = r.data || []
}

async function closeT(id) {
  if (!confirm('Закрыть заявление?')) return
  await adminApi.closeTicket(id)
  await load()
}

async function deleteT(id) {
  if (!confirm('Удалить заявление безвозвратно?')) return
  await adminApi.deleteTicket(id)
  await load()
}

async function doExportXLSX() {
  const r = await adminApi.listTickets(filter)
  exportXLSX(r.data || [])
}

function exportXLSX(data) {
  import('xlsx').then(XLSX => {
    const ws = XLSX.utils.json_to_sheet(data.map(t => ({
      '№': t.id, 'Имя': t.first_name, 'Фамилия': t.last_name,
      'Подразделение': t.division, 'Описание': t.description,
      'Приоритет': plabel(t.priority), 'Статус': slabel(t.status),
      'Дата': new Date(t.created_at).toLocaleString('ru-RU'),
      'Комментарий ИТ': t.admin_comment,
    })))
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, 'Заявления')
    XLSX.writeFile(wb, 'tickets.xlsx')
  })
}

function count(field, val) { return tickets.value.filter(t => t[field] === val).length }
function fmt(d)     { return new Date(d).toLocaleDateString('ru-RU') }
function fmtTime(d) { return new Date(d).toLocaleTimeString('ru-RU', { hour:'2-digit', minute:'2-digit' }) }
function plabel(p)  { return { high:'Высокий', medium:'Средний', low:'Низкий' }[p] || p }
function slabel(s)  { return { open:'Открыто', in_progress:'На рассмотрении', closed:'Закрыто', cancelled:'Отменено' }[s] || s }

let pollInterval     = null
let chatPollInterval = null

onMounted(async () => {
  await load()
  loadIPLogs()

  // Polling 1: каждые 10 сек обновляем список заявок + непрочитанные
  pollInterval = setInterval(async () => {
    try {
      const params = {}
      Object.entries(filter).forEach(([k,v]) => { if(v) params[k] = v })
      const [r, u] = await Promise.all([
        adminApi.listTickets(params),
        adminApi.unreadCounts()
      ])
      tickets.value = r.data || []
      // Обновляем unread точечно — не трогаем уже убранные
      const fresh = u.data || {}
      const merged = { ...unread.value }
      Object.keys(fresh).forEach(id => { merged[id] = fresh[id] })
      unread.value = merged
    } catch(e) { /* тихо */ }
  }, 10000)
})

onUnmounted(() => {
  clearInterval(pollInterval)
  clearInterval(chatPollInterval)
})

function startChatPolling() {
  stopChatPolling()
  chatPollInterval = setInterval(async () => {
    if (!activeTicket.value || !chatModal.value) {
      stopChatPolling()
      return
    }
    try {
      const r = await adminApi.getTicket(activeTicket.value.id)
      if (r.data?.messages) {
        activeTicket.value = { ...activeTicket.value, messages: r.data.messages }
      }
    } catch(e) { /* тихо */ }
  }, 5000)
}

function stopChatPolling() {
  clearInterval(chatPollInterval)
  chatPollInterval = null
}
</script>

<style scoped>
.admin-header {
  background: var(--grad-dark); border-radius: var(--radius);
  padding: 22px 28px; margin-bottom: 24px; color: white;
  display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 14px;
}
.admin-header h1 { font-family: 'Playfair Display', serif; font-size: 22px; }
.admin-header p  { font-size: 12px; opacity: .55; margin-top: 3px; }
.ip-info { text-align: right; }
.ip-tag  { background: rgba(79,124,255,.22); border: 1px solid rgba(79,124,255,.4); color: #a0b4ff; font-size: 11px; padding: 4px 10px; border-radius: 8px; }
.ip-time { font-size: 11px; opacity: .45; margin-top: 4px; }

.filter-bar { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 16px; }
.filter-bar select, .filter-bar input {
  padding: 8px 12px; border-radius: var(--radius-sm); border: 1.5px solid var(--border);
  background: var(--surface); font-family: 'Onest', sans-serif; font-size: 12px; outline: none; color: var(--text);
}
.filter-bar select:focus, .filter-bar input:focus { border-color: var(--accent); }

.export-row { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 18px; align-items: center; }

.desc-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.stat-card.clickable {
  cursor: pointer;
  transition: all .2s;
  user-select: none;
}
.stat-card.clickable:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
  border-color: var(--accent);
}
.stat-card.active {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(79,124,255,.2);
}

/* Modal  */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 999; padding: 20px;
}
.modal-box {
  max-width: 760px; width: 100%;
  max-height: 90vh; overflow-y: auto;
}
.modal-top {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 12px; gap: 12px;
}
.ticket-info-row {
  display: flex; flex-wrap: wrap; gap: 10px;
  font-size: 12px; color: var(--text-muted); margin-top: 6px;
}
.ticket-desc {
  background: var(--surface2); border-radius: 10px;
  padding: 12px; font-size: 13px; margin-bottom: 12px;
  color: var(--text); line-height: 1.5;
}
.btn-with-badge { position: relative; display: inline-flex; }
.unread-badge {
  position: absolute; top: -6px; right: -6px;
  background: #ef4444; color: white;
  font-size: 10px; font-weight: 700;
  min-width: 18px; height: 18px;
  border-radius: 9px; display: flex;
  align-items: center; justify-content: center;
  padding: 0 4px; border: 2px solid white;
  pointer-events: none;
  z-index: 1;
}
</style>