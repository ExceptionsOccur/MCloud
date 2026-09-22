import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as apiLogin, getUserInfo } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)

  async function login(username, password) {
    const res = await apiLogin(username, password)
    if (res.code === 0) {
      token.value = res.data.access_token
      user.value = res.data.user
      localStorage.setItem('token', res.data.access_token)
      return true
    }
    return false
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  async function fetchUserInfo() {
    if (!token.value) return
    try {
      const res = await getUserInfo()
      if (res.code === 0) {
        user.value = res.data
      } else {
        logout()
      }
    } catch {
      logout()
    }
  }

  return { token, user, login, logout, fetchUserInfo }
})
