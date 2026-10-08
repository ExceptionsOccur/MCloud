import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import Login from '../views/Login.vue'
import HostManagement from '../views/HostManagement.vue'
import ResourceStatistics from '../views/ResourceStatistics.vue'
import IpStatistics from '../views/IpStatistics.vue'
import BusinessStatistics from '../views/BusinessStatistics.vue'
import PersonnelManagement from '../views/PersonnelManagement.vue'
import PublicIPManagement from '../views/PublicIPManagement.vue'
import ZeroTrustLedger from '../views/ZeroTrustLedger.vue'
import MappingLedger from '../views/MappingLedger.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'HostManagement',
    component: HostManagement,
    meta: { requiresAuth: true }
  },
  {
    path: '/statistics',
    name: 'ResourceStatistics',
    component: ResourceStatistics,
    meta: { requiresAuth: true }
  },
  {
    path: '/ip-statistics',
    name: 'IpStatistics',
    component: IpStatistics,
    meta: { requiresAuth: true }
  },
  {
    path: '/business-statistics',
    name: 'BusinessStatistics',
    component: BusinessStatistics,
    meta: { requiresAuth: true }
  },
  {
    path: '/personnel',
    name: 'PersonnelManagement',
    component: PersonnelManagement,
    meta: { requiresAuth: true }
  },
  {
    path: '/public-ip',
    name: 'PublicIPManagement',
    component: PublicIPManagement,
    meta: { requiresAuth: true }
  },
  {
    path: '/zero-trust',
    name: 'ZeroTrustLedger',
    component: ZeroTrustLedger,
    meta: { requiresAuth: true }
  },
  {
    path: '/mapping-ledger',
    name: 'MappingLedger',
    component: MappingLedger,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.token) {
    next('/login')
  } else if (to.path === '/login' && authStore.token) {
    next('/')
  } else {
    next()
  }
})

export default router
