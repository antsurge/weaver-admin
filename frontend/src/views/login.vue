<template>
    <div>
        <AuthToolbar />
        你好
        {{ $t("common.hello") }}
        <!-- 背景 -->
        <!-- <div id="bubble" class="bubble">
            <canvas id="bubble-canvas" class="bubble-canvas"></canvas>
        </div> -->
    </div>
</template>

<script setup lang="ts">
import { $t } from "@/locales"
// 引入国际化
import AuthToolbar from "@/layouts/auth/toolbar.vue"
</script>

<style scoped lang="scss"></style>