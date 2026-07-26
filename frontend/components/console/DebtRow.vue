<script setup lang="ts">
const props = defineProps<{
  debt: any
  selectedDebts: number[]
  paymentAmount: number
  expanded: boolean
  payments: any[]
  money: (value: number) => string
  formatNumber: (value: number) => string
}>()

const emit = defineEmits<{
  'update:selectedDebts': [value: number[]]
  'update:paymentAmount': [value: number]
  pay: []
  toggleHistory: []
  reverse: [payment: any]
}>()

function toggleSelection(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  const id = Number((event.target as HTMLInputElement).value)
  emit('update:selectedDebts', checked ? [...new Set([...props.selectedDebts, id])] : props.selectedDebts.filter((value: number) => value !== id))
}

function updateAmount(event: Event) {
  const value = Number((event.target as HTMLInputElement).value.replace(/\D/g, ''))
  emit('update:paymentAmount', value)
}
</script>

<template>
  <article class="debt">
    <input
      v-if="debt.status === 'OPEN'"
      class="debt-select"
      type="checkbox"
      :value="Number(debt.id)"
      :checked="selectedDebts.includes(Number(debt.id))"
      :aria-label="`Pilih piutang ${debt.invoice_no}`"
      @change="toggleSelection"
    >
    <span v-else class="debt-select-spacer" />

    <div class="debt-info">
      <strong>{{ debt.customer_name }}</strong>
      <small>Invoice {{ debt.invoice_no }}</small>
      <small>Awal {{ money(debt.original_amount) }} · Sisa <b>{{ money(debt.remaining_amount) }}</b></small>
    </div>

    <div class="debt-actions">
      <form v-if="debt.status === 'OPEN'" @submit.prevent="emit('pay')">
        <label>
          Nominal Pembayaran (Rp)
          <input
            :value="formatNumber(paymentAmount)"
            inputmode="numeric"
            pattern="[0-9.]+"
            placeholder="0"
            required
            @input="updateAmount"
          >
        </label>
        <button class="primary" :disabled="!paymentAmount || paymentAmount > debt.remaining_amount">Catat Pembayaran</button>
      </form>
      <span v-else class="paid">Lunas</span>
      <button type="button" class="soft" @click="emit('toggleHistory')">
        {{ expanded ? 'Tutup Histori' : 'Histori Pembayaran' }}
      </button>
    </div>

    <section v-if="expanded" class="debt-payment-history">
      <article v-for="payment in payments" :key="payment.id" :class="{ reversed: payment.reversed_at }">
        <div>
          <strong>{{ money(payment.amount) }}</strong>
          <small>{{ new Date(payment.payment_date).toLocaleString('id-ID') }}</small>
          <small v-if="payment.notes">{{ payment.notes }}</small>
        </div>
        <div v-if="payment.reversed_at" class="reversal-detail">
          <span class="reversed-badge">Dikoreksi</span>
          <small>{{ new Date(payment.reversed_at).toLocaleString('id-ID') }} · {{ payment.reversed_by_name || 'Pengguna' }}</small>
          <small>{{ payment.reversal_reason }}</small>
        </div>
        <button v-else type="button" class="danger" @click="emit('reverse', payment)">Koreksi</button>
      </article>
      <p v-if="!payments.length" class="empty">Belum ada pembayaran yang dicatat.</p>
    </section>
  </article>
</template>
