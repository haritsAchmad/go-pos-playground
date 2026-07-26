<script setup lang="ts">
defineProps<{ item: any, movements: any[] }>()
defineEmits<{ close: [] }>()
const labels:Record<string,string>={OPENING_BALANCE:'Saldo awal histori',INITIAL_STOCK:'Stok awal',MANUAL_ADJUSTMENT:'Penyesuaian manual',SALE:'Penjualan',PURCHASE:'Pembelian',TRANSACTION_EDIT_RESTORE:'Edit transaksi · stok lama dikembalikan',TRANSACTION_EDIT_APPLY:'Edit transaksi · stok baru diterapkan',TRANSACTION_VOID:'Pembatalan transaksi',BULK_RESET:'Reset stok massal'}
</script>

<template>
  <div class="modal-backdrop" @mousedown.self="$emit('close')">
    <section class="modal-card stock-history-modal" role="dialog" aria-modal="true">
      <div class="modal-head"><div><h2>Histori Stok</h2><p>{{item.sku}} · {{item.name}}</p></div><button class="modal-close" aria-label="Tutup" @click="$emit('close')">×</button></div>
      <div class="stock-history-summary">Stok saat ini <strong>{{item.stock_display}}</strong></div>
      <div class="stock-history-list">
        <article v-for="movement in movements" :key="movement.id">
          <div><strong>{{labels[movement.movement_type] || movement.movement_type}}</strong><small>{{new Date(movement.created_at).toLocaleString('id-ID')}}</small><small v-if="movement.notes">{{movement.notes}}</small></div>
          <div class="stock-change" :class="movement.quantity_change<0?'out':'in'"><b>{{movement.quantity_change>0?'+':''}}{{movement.quantity_change}}</b><small>{{movement.quantity_before}} → {{movement.quantity_after}}</small></div>
        </article>
        <p v-if="!movements.length" class="empty">Belum ada pergerakan stok yang tercatat.</p>
      </div>
      <div class="modal-actions"><button class="soft" @click="$emit('close')">Tutup</button></div>
    </section>
  </div>
</template>
