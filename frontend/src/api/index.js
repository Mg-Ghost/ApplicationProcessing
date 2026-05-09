import axios from 'axios'

const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use(cfg => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

api.interceptors.response.use(
  r => r,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

// ─── Auth ─────────────────────────────────────────────────────────────────
export const authApi = {
  register: d => api.post('/auth/register', d),
  login:    d => api.post('/auth/login', d),
  adminLogin: d => api.post('/auth/admin/login', d),
}

// ─── Profile ──────────────────────────────────────────────────────────────
export const profileApi = {
  get:    () => api.get('/user/profile'),
  update: d  => api.put('/user/profile', d),
}

// ─── Tickets (user) ───────────────────────────────────────────────────────
export const ticketsApi = {
  list:      ()        => api.get('/tickets'),
  get:       id        => api.get(`/tickets/${id}`),
  create:    d         => api.post('/tickets', d),
  update:    (id, d)   => api.put(`/tickets/${id}`, d),
  cancel:    id        => api.patch(`/tickets/${id}/cancel`),
  close:     id        => api.patch(`/tickets/${id}/close`),
  reply:     (id, d)   => api.post(`/tickets/${id}/reply`, d),  // ответ пользователя
}

// ─── Admin ────────────────────────────────────────────────────────────────
export const adminApi = {
  listTickets:  params  => api.get('/admin/tickets', { params }),
  getTicket:    id      => api.get(`/admin/tickets/${id}`),
  deleteTicket: id      => api.delete(`/admin/tickets/${id}`),
  closeTicket:  id      => api.patch(`/admin/tickets/${id}/close`),
  addComment:   (id, d) => api.post(`/admin/tickets/${id}/comment`, d),
  export:       params  => api.get('/admin/tickets/export', { params }),
  ipLogs:       ()      => api.get('/admin/ip-logs'),
}

export default api
