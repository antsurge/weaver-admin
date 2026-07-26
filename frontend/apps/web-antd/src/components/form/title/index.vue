<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <div class="form-title">
    <div class="form-title-bar"></div>
    <div class="form-title-text">
      {{ title }}
    </div>
  </div>
</template>

<style scoped>
.form-title {
  display: flex;
  align-items: center;
  width: 100%;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.form-title-bar {
  width: 4px;
  height: 16px;
  background: hsl(var(--primary));
  border-radius: 4px;
  margin-right: 6px;
}

.form-title-text {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}
</style>
