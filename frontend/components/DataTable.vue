<script setup lang="ts">
defineProps<{rows:any[],columns:any[][],actions?:boolean,voidActions?:boolean}>()
defineEmits(['edit','delete','void'])
const money=(v:number)=>new Intl.NumberFormat('id-ID',{style:'currency',currency:'IDR',maximumFractionDigits:0}).format(v||0)
const display=(row:any,col:any[])=>col[2]==='money'?money(row[col[0]]):col[2]==='date'?new Date(row[col[0]]).toLocaleString('id-ID'):row[col[0]]??'—'
</script>
<template><section class="panel"><div class="table-wrap"><table><thead><tr><th v-for="c in columns" :class="c[3]">{{c[1]}}</th><th v-if="actions||voidActions">Aksi</th></tr></thead><tbody><tr v-for="row in rows" :key="row.id" :class="{voided:row.status==='VOID'}"><td v-for="c in columns" :class="c[3]">{{display(row,c)}}</td><td v-if="actions" class="table-actions"><button class="soft" :disabled="row.code==='UMUM'" @click="$emit('edit',row)">Ubah</button><button class="danger" :disabled="row.code==='UMUM'" @click="$emit('delete',row)">Hapus</button></td><td v-if="voidActions"><button class="danger" :disabled="row.status==='VOID'" :title="row.void_reason||''" @click="$emit('void',row)">{{row.status==='VOID'?'Dibatalkan':'Batalkan'}}</button></td></tr><tr v-if="!rows.length"><td :colspan="columns.length+((actions||voidActions)?1:0)" class="empty">Belum ada data</td></tr></tbody></table></div></section></template>
