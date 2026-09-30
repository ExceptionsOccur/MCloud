import api from './index'

export function getPersons(keyword = '') {
  return api.get('/persons', { params: { keyword } })
}

export function createPerson(data) {
  return api.post('/persons', data)
}

export function updatePerson(id, data) {
  return api.put(`/persons/${id}`, data)
}

export function deletePerson(id) {
  return api.delete(`/persons/${id}`)
}
