<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { inventoryApi } from './api'

const activeTab = ref('items')
const items = ref([])
const suppliers = ref([])
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const editingId = ref(null)

const itemForm = reactive({ supplier_id: '', name: '', description: '', stock: 0 })
const supplierForm = reactive({ name: '', phone: '', address: '' })

const title = computed(() => activeTab.value === 'items' ? 'Items' : 'Suppliers')
const records = computed(() => activeTab.value === 'items' ? items.value : suppliers.value)

function clearMessages() {
  errorMessage.value = ''
  successMessage.value = ''
}

function resetForm() {
  editingId.value = null
  Object.assign(itemForm, { supplier_id: '', name: '', description: '', stock: 0 })
  Object.assign(supplierForm, { name: '', phone: '', address: '' })
}

async function loadData() {
  loading.value = true
  clearMessages()
  try {
    const [itemData, supplierData] = await Promise.all([
      inventoryApi.getItems(),
      inventoryApi.getSuppliers(),
    ])
    items.value = Array.isArray(itemData) ? itemData : []
    suppliers.value = Array.isArray(supplierData) ? supplierData : []
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

function selectTab(tab) {
  activeTab.value = tab
  resetForm()
  clearMessages()
}

function editRecord(record) {
  editingId.value = record.id
  clearMessages()
  if (activeTab.value === 'items') {
    Object.assign(itemForm, {
      supplier_id: record.supplier_id ?? '',
      name: record.name ?? '',
      description: record.description ?? '',
      stock: record.stock ?? 0,
    })
  } else {
    Object.assign(supplierForm, {
      name: record.name ?? '',
      phone: record.phone ?? '',
      address: record.address ?? '',
    })
  }
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function submitForm() {
  clearMessages()
  saving.value = true
  try {
    if (activeTab.value === 'items') {
      const payload = {
        supplier_id: Number(itemForm.supplier_id),
        name: itemForm.name.trim(),
        description: itemForm.description.trim(),
        stock: Number(itemForm.stock),
      }
      if (!payload.supplier_id || !payload.name || payload.stock < 0) {
        throw new Error('Supplier, item name, and non-negative stock are required.')
      }
      editingId.value
        ? await inventoryApi.updateItem(editingId.value, payload)
        : await inventoryApi.createItem(payload)
    } else {
      const payload = {
        name: supplierForm.name.trim(),
        phone: supplierForm.phone.trim(),
        address: supplierForm.address.trim(),
      }
      if (!payload.name || !payload.phone || !payload.address) {
        throw new Error('Name, phone, and address are required.')
      }
      editingId.value
        ? await inventoryApi.updateSupplier(editingId.value, payload)
        : await inventoryApi.createSupplier(payload)
    }

    successMessage.value = `${title.value.slice(0, -1)} ${editingId.value ? 'updated' : 'created'} successfully.`
    resetForm()
    await loadData()
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    saving.value = false
  }
}

async function removeRecord(record) {
  if (!window.confirm(`Delete ${record.name}?`)) return
  clearMessages()
  try {
    activeTab.value === 'items'
      ? await inventoryApi.deleteItem(record.id)
      : await inventoryApi.deleteSupplier(record.id)
    successMessage.value = `${title.value.slice(0, -1)} deleted successfully.`
    if (editingId.value === record.id) resetForm()
    await loadData()
  } catch (error) {
    errorMessage.value = error.message
  }
}

function supplierName(id) {
  return suppliers.value.find((supplier) => supplier.id === id)?.name || `Supplier #${id}`
}

onMounted(loadData)
</script>

<template>
  <main class="shell">
    <header class="hero">
      <div>
        <p class="eyebrow">GO + POSTGRESQL PLAYGROUND</p>
        <h1>Inventory Console</h1>
        <p class="subtitle">A small Vue client for experimenting with the inventory REST API.</p>
      </div>
      <button class="secondary" :disabled="loading" @click="loadData">Refresh</button>
    </header>

    <nav class="tabs" aria-label="Inventory sections">
      <button :class="{ active: activeTab === 'items' }" @click="selectTab('items')">Items</button>
      <button :class="{ active: activeTab === 'suppliers' }" @click="selectTab('suppliers')">Suppliers</button>
    </nav>

    <p v-if="errorMessage" class="alert error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="alert success">{{ successMessage }}</p>

    <section class="panel form-panel">
      <div class="section-heading">
        <div>
          <p class="eyebrow">{{ editingId ? 'EDIT RECORD' : 'NEW RECORD' }}</p>
          <h2>{{ editingId ? `Update ${title.slice(0, -1)}` : `Add ${title.slice(0, -1)}` }}</h2>
        </div>
        <button v-if="editingId" class="ghost" @click="resetForm">Cancel edit</button>
      </div>

      <form v-if="activeTab === 'items'" class="form-grid" @submit.prevent="submitForm">
        <label>
          Supplier
          <select v-model="itemForm.supplier_id" required>
            <option value="" disabled>Select supplier</option>
            <option v-for="supplier in suppliers" :key="supplier.id" :value="supplier.id">
              {{ supplier.name }}
            </option>
          </select>
        </label>
        <label>
          Item name
          <input v-model="itemForm.name" minlength="3" maxlength="100" required />
        </label>
        <label>
          Stock
          <input v-model.number="itemForm.stock" type="number" min="0" required />
        </label>
        <label class="wide">
          Description
          <textarea v-model="itemForm.description" rows="3" />
        </label>
        <button class="primary" :disabled="saving || suppliers.length === 0">
          {{ saving ? 'Saving…' : editingId ? 'Save changes' : 'Create item' }}
        </button>
      </form>

      <form v-else class="form-grid" @submit.prevent="submitForm">
        <label>
          Supplier name
          <input v-model="supplierForm.name" minlength="3" maxlength="100" required />
        </label>
        <label>
          Phone
          <input v-model="supplierForm.phone" required />
        </label>
        <label class="wide">
          Address
          <textarea v-model="supplierForm.address" rows="3" required />
        </label>
        <button class="primary" :disabled="saving">
          {{ saving ? 'Saving…' : editingId ? 'Save changes' : 'Create supplier' }}
        </button>
      </form>
    </section>

    <section class="panel">
      <div class="section-heading">
        <div>
          <p class="eyebrow">CURRENT DATA</p>
          <h2>{{ title }} <span class="count">{{ records.length }}</span></h2>
        </div>
      </div>

      <p v-if="loading" class="empty">Loading data…</p>
      <p v-else-if="records.length === 0" class="empty">No {{ title.toLowerCase() }} found.</p>

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr v-if="activeTab === 'items'">
              <th>ID</th><th>Name</th><th>Supplier</th><th>Stock</th><th>Description</th><th>Actions</th>
            </tr>
            <tr v-else>
              <th>ID</th><th>Name</th><th>Phone</th><th>Address</th><th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in records" :key="record.id">
              <template v-if="activeTab === 'items'">
                <td>#{{ record.id }}</td>
                <td class="strong">{{ record.name }}</td>
                <td>{{ supplierName(record.supplier_id) }}</td>
                <td><span class="stock">{{ record.stock }}</span></td>
                <td>{{ record.description || '—' }}</td>
              </template>
              <template v-else>
                <td>#{{ record.id }}</td>
                <td class="strong">{{ record.name }}</td>
                <td>{{ record.phone }}</td>
                <td>{{ record.address }}</td>
              </template>
              <td class="actions">
                <button class="ghost" @click="editRecord(record)">Edit</button>
                <button class="danger" @click="removeRecord(record)">Delete</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </main>
</template>
