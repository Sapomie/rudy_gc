<template>
  <div class="loading-state">
    <div class="loading-line"></div>
    <div class="loading-line short"></div>
    <div class="loading-grid">
      <span v-for="item in 12" :key="item"></span>
    </div>
  </div>
</template>
