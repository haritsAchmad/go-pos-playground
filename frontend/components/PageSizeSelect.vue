<script setup lang="ts">
const model = defineModel<number>({ required: true })
</script>

<template>
  <label class="page-size-select">
    <span>Baris:</span>
    <select v-model.number="model">
      <option :value="10">10</option>
      <option :value="20">20</option>
      <option :value="50">50</option>
    </select>
  </label>
</template>
