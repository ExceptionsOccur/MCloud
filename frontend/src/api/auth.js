import api from './index'

export function login(username, password) {
  return api.post('/auth/login', { username, password })
}

export function getUserInfo() {
  return api.get('/auth/me')
}

export function changePassword(oldPassword, newPassword) {
  return api.post('/auth/password', { old_password: oldPassword, new_password: newPassword })
}
