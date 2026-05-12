<template>
  <div class="page">
    <div v-if="loading" class="card" style="max-width:800px;">Загрузка...</div>

    <template v-else-if="ticket">
      <!-- Header -->
      <div class="page-header" style="display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:12px;">
        <div>
          <h1>Заявление №{{ ticket.id }}</h1>
          <p>Создано {{ fmt(ticket.created_at) }} · {{ ticket.division }}</p>
        </div>
        <div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center;">
          <!-- Приоритет видит только администратор -->
          <template v-if="isAdmin">
            <span :class="['badge', `badge-${ticket.priority}`]" style="font-size:13px;padding:5px 12px;">
              {{ plabel(ticket.priority) }}
              <span v-if="ticket.auto_escalated" title="Автоматически повышен">⚡</span>
            </span>
          </template>
          <template v-else>
            <span v-if="ticket.auto_escalated" class="badge badge-high" style="font-size:13px;padding:5px 12px;">
              Высокий приоритет ⚡
            </span>
          </template>
          <span :class="['badge', `badge-${ticket.status}`]" style="font-size:13px;padding:5px 12px;">
            {{ slabel(ticket.status) }}
          </span>
        </div>
      </div>

      <div class="detail-grid">

        <!-- Левая колонка: основная информация -->
        <div style="display:flex;flex-direction:column;gap:16px;">

          <!-- Заявитель -->
          <div class="card">
            <div class="card-title" style="font-size:15px;">👤 Заявитель</div>
            <div class="info-grid">
              <div class="info-row"><span class="info-label">Имя</span><span>{{ ticket.first_name }} {{ ticket.last_name }}</span></div>
              <div class="info-row"><span class="info-label">Телефон</span><span>{{ ticket.phone || '—' }}</span></div>
              <div class="info-row"><span class="info-label">Должность</span><span>{{ ticket.position || '—' }}</span></div>
              <div class="info-row"><span class="info-label">Подразделение</span><span>{{ ticket.division }}</span></div>
              <div class="info-row"><span class="info-label">Кабинет</span><span>{{ ticket.room || '—' }}</span></div>
            </div>
          </div>

          <!-- Описание проблемы -->
          <div class="card">
            <div class="card-title" style="font-size:15px;">🖥 Описание проблемы</div>
            <p style="font-size:14px;line-height:1.7;color:var(--text);">{{ ticket.description }}</p>
            <div class="divider"></div>
            <div class="info-grid">
              <div class="info-row">
                <span class="info-label">Инвентарный №</span>
                <span>{{ ticket.inventory_number || '—' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">IP-адрес</span>
                <code v-if="ticket.ip_address" style="font-size:13px;background:var(--surface2);padding:2px 7px;border-radius:5px;">{{ ticket.ip_address }}</code>
                <span v-else>—</span>
              </div>
            </div>
          </div>

          <!-- Даты -->
          <div class="card">
            <div class="card-title" style="font-size:15px;">📅 Хронология</div>
            <div class="info-grid">
              <div class="info-row"><span class="info-label">Создано</span><span>{{ fmtFull(ticket.created_at) }}</span></div>
              <div class="info-row"><span class="info-label">Обновлено</span><span>{{ fmtFull(ticket.updated_at) }}</span></div>
              <div v-if="isAdmin && ticket.closed_by_admin" class="info-row">
                <span class="info-label">Закрыл</span>
                <span :class="ticket.closed_by_role === 'admin' ? 'closed-by-admin' : 'closed-by-user'">
                  {{ ticket.closed_by_role === 'admin' ? '🔧 Администратор:' : '👤 Пользователь:' }}
                  <strong>{{ ticket.closed_by_admin }}</strong>
                </span>
              </div>
            </div>
          </div>

          <!-- Кнопки действий -->
          <div style="display:flex;gap:10px;flex-wrap:wrap;">
            <button class="btn btn-ghost" @click="$router.back()">← Назад</button>

            <!-- Кнопки пользователя -->
            <template v-if="!isAdmin">
              <router-link
                v-if="ticket.status === 'open' || ticket.status === 'in_progress'"
                :to="`/tickets/${ticket.id}/edit`"
                class="btn btn-primary">✏️ Редактировать</router-link>
              <button
                v-if="ticket.status === 'open'"
                class="btn btn-warn"
                @click="doCancel">Отменить</button>
              <button
                v-if="ticket.status === 'open' || ticket.status === 'in_progress'"
                class="btn btn-ghost"
                @click="doClose">✓ Закрыть</button>
            </template>

            <!-- Кнопки администратора -->
            <template v-if="isAdmin">
              <button
                v-if="ticket.status !== 'closed'"
                class="btn btn-success"
                @click="doAdminClose">✓ Закрыть заявление</button>
              <button
                class="btn btn-danger"
                @click="doDelete">🗑 Удалить</button>
            </template>
          </div>
        </div>

        <!-- Правая колонка: переписка -->
        <div class="card" style="padding:0;overflow:hidden;align-self:start;">
          <TicketChat
            :ticket-id="ticket.id"
            :messages="ticket.messages || []"
            :status="ticket.status"
            :is-admin="isAdmin"
            @sent="reload"
          />
        </div>

      </div>
    </template>

    <div v-else class="card">Заявление не найдено.</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ticketsApi, adminApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import TicketChat from '@/components/shared/TicketChat.vue'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()
const id     = route.params.id

const ticket  = ref(null)
const loading = ref(true)

const isAdmin = computed(() => auth.isAdmin)

async function reload() {
  try {
    const r = isAdmin.value
      ? await adminApi.getTicket(id)
      : await ticketsApi.get(id)
    ticket.value = r.data
  } catch(e) {
    ticket.value = null
  }
}

onMounted(async () => {
  await reload()
  loading.value = false
})

async function doCancel() {
  if (!confirm('Отменить заявление?')) return
  await ticketsApi.cancel(id)
  router.push('/dashboard')
}

async function doClose() {
  if (!confirm('Закрыть заявление (проблема решена)?')) return
  await ticketsApi.close(id)
  await reload()
}

async function doAdminClose() {
  if (!confirm('Закрыть заявление?')) return
  await adminApi.closeTicket(id)
  await reload()
}

async function doDelete() {
  if (!confirm('Удалить заявление безвозвратно?')) return
  await adminApi.deleteTicket(id)
  router.push('/admin')
}

function fmt(d)     { return new Date(d).toLocaleDateString('ru-RU') }
function fmtFull(d) { return new Date(d).toLocaleString('ru-RU', { day:'2-digit', month:'2-digit', year:'numeric', hour:'2-digit', minute:'2-digit' }) }
function plabel(p)  { return { high:'Высокий', medium:'Средний', low:'Низкий' }[p] || p }
function slabel(s)  { return { open:'Открыто', in_progress:'На рассмотрении', closed:'Закрыто', cancelled:'Отменено' }[s] || s }
</script>

<style scoped>
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 20px;
  max-width: 1100px;
  align-items: start;
}
@media (max-width: 900px) {
  .detail-grid { grid-template-columns: 1fr; }
}

.info-grid { display: flex; flex-direction: column; gap: 10px; }
.info-row  { display: flex; gap: 12px; font-size: 13px; align-items: baseline; }
.info-label {
  min-width: 130px; color: var(--text-muted);
  font-size: 11px; font-weight: 600;
  text-transform: uppercase; letter-spacing: .4px;
}

.btn-success { background: var(--grad-success); color: white; }
.closed-by-admin { color: var(--accent); font-weight: 600; font-size: 13px; }
.closed-by-user  { color: #16a34a;       font-weight: 600; font-size: 13px; }
</style>
