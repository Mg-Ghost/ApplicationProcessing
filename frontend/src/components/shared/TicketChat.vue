<template>
  <div class="chat-wrap">
    <div class="chat-header">
      <span>💬 Переписка по заявлению #{{ ticketId }}</span>
      <span v-if="msgs.length" class="chat-count">{{ msgs.length }} сообщ.</span>
      <button class="refresh-btn" @click="fetchMessages" title="Обновить">↻</button>
    </div>

    <!-- Лента сообщений -->
    <div class="chat-feed" ref="feedEl">
      <div v-if="loadingMsgs" class="chat-empty">Загрузка...</div>
      <div v-else-if="!msgs.length" class="chat-empty">
        Сообщений пока нет. {{ isAdmin ? 'Напишите ответ пользователю.' : 'Здесь появятся ответы IT-отдела.' }}
      </div>
      <div
        v-for="m in msgs"
        :key="m.id"
        :class="['msg', m.author === 'admin' ? 'msg-admin' : 'msg-user']"
      >
        <div class="msg-meta">
          <span class="msg-author">{{ m.author === 'admin' ? '🔧 ' : '👤 ' }}{{ m.author_name }}</span>
          <span class="msg-time">{{ fmtTime(m.created_at) }}</span>
        </div>
        <div class="msg-text">{{ m.text }}</div>
      </div>
    </div>

    <!-- Поле ввода -->
    <div v-if="canReply" class="chat-input-row">
      <textarea
        v-model="draft"
        :placeholder="isAdmin ? 'Ответ от IT-отдела... (Ctrl+Enter)' : 'Ваш вопрос или уточнение... (Ctrl+Enter)'"
        @keydown.ctrl.enter="send"
        rows="2"
      ></textarea>
      <button class="btn btn-primary btn-sm send-btn" :disabled="!draft.trim() || sending" @click="send">
        {{ sending ? '...' : 'Отправить' }}
      </button>
    </div>
    <div v-else class="chat-closed-note">
      Переписка закрыта
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { ticketsApi, adminApi } from '@/api'

const props = defineProps({
  ticketId: { type: [Number, String], required: true },
  status:   { type: String, default: 'open' },
  isAdmin:  { type: Boolean, default: false },
})

const emit = defineEmits(['sent'])

const msgs       = ref([])
const draft      = ref('')
const sending    = ref(false)
const loadingMsgs = ref(false)
const feedEl     = ref(null)

const canReply = computed(() =>
  props.status !== 'closed' && props.status !== 'cancelled'
)

// Загружаем сообщения самостоятельно через API
async function fetchMessages() {
  if (!props.ticketId) return
  loadingMsgs.value = true
  try {
    let r
    if (props.isAdmin) {
      r = await adminApi.getTicket(props.ticketId)
    } else {
      r = await ticketsApi.get(props.ticketId)
    }
    msgs.value = r.data?.messages || []
    await scrollDown()
  } catch(e) {
    console.error('fetchMessages error:', e)
  } finally {
    loadingMsgs.value = false
  }
}

async function send() {
  if (!draft.value.trim()) return
  sending.value = true
  try {
    if (props.isAdmin) {
      await adminApi.addComment(props.ticketId, { comment: draft.value.trim() })
    } else {
      await ticketsApi.reply(props.ticketId, { text: draft.value.trim() })
    }
    draft.value = ''
    await fetchMessages()  // перезагружаем сразу после отправки
    emit('sent')
  } catch(e) {
    alert('Ошибка отправки: ' + (e.response?.data?.error || e.message))
  } finally {
    sending.value = false
  }
}

async function scrollDown() {
  await nextTick()
  if (feedEl.value) feedEl.value.scrollTop = feedEl.value.scrollHeight
}

function fmtTime(d) {
  return new Date(d).toLocaleString('ru-RU', {
    day: '2-digit', month: '2-digit', year: '2-digit',
    hour: '2-digit', minute: '2-digit'
  })
}

// Загружаем при монтировании и при смене ticketId
onMounted(fetchMessages)
watch(() => props.ticketId, fetchMessages)
</script>

<style scoped>
.chat-wrap {
  border: 1.5px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--surface);
}
.chat-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; background: var(--surface2);
  border-bottom: 1px solid var(--border);
  font-size: 13px; font-weight: 600; color: var(--text); gap: 8px;
}
.chat-count {
  background: var(--grad-btn); color: white;
  font-size: 10px; padding: 2px 8px; border-radius: 20px; margin-left: auto;
}
.refresh-btn {
  background: none; border: none; cursor: pointer;
  font-size: 15px; color: var(--text-muted); padding: 0 4px;
  transition: color .2s;
}
.refresh-btn:hover { color: var(--accent); }

.chat-feed {
  min-height: 140px; max-height: 340px;
  overflow-y: auto; padding: 14px;
  display: flex; flex-direction: column; gap: 10px;
}
.chat-empty {
  color: var(--text-muted); font-size: 13px;
  text-align: center; padding: 24px 0;
}

.msg { max-width: 80%; display: flex; flex-direction: column; gap: 3px; }
.msg-admin { align-self: flex-start; }
.msg-user  { align-self: flex-end; }

.msg-meta { display: flex; gap: 8px; align-items: baseline; }
.msg-author { font-size: 11px; font-weight: 600; color: var(--text-muted); }
.msg-time   { font-size: 10px; color: var(--text-muted); }

.msg-admin .msg-text {
  background: var(--surface2); border: 1px solid var(--border);
  border-radius: 4px 14px 14px 14px;
  padding: 9px 13px; font-size: 13px; line-height: 1.5;
  border-left: 3px solid var(--accent);
}
.msg-user .msg-text {
  background: var(--grad-btn); color: white;
  border-radius: 14px 4px 14px 14px;
  padding: 9px 13px; font-size: 13px; line-height: 1.5;
}

.chat-input-row {
  display: flex; gap: 8px; padding: 10px 14px;
  border-top: 1px solid var(--border); background: var(--surface2);
  align-items: flex-end;
}
.chat-input-row textarea {
  flex: 1; padding: 8px 11px; border-radius: 9px;
  border: 1.5px solid var(--border); background: white;
  font-family: 'Onest', sans-serif; font-size: 13px;
  outline: none; resize: none; transition: border-color .2s;
}
.chat-input-row textarea:focus { border-color: var(--accent); }
.send-btn { align-self: flex-end; white-space: nowrap; }

.chat-closed-note {
  padding: 10px 14px; font-size: 12px; color: var(--text-muted);
  border-top: 1px solid var(--border); text-align: center;
  background: var(--surface2);
}
</style>