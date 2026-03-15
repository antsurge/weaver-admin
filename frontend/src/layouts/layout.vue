<!-- src/layouts/AdminLayout.vue -->
<template>
    <a-layout style="min-height: 100vh">
      <!-- 左侧侧边栏 -->
      <a-layout-sider
        :collapsed="collapsed"
        collapsible
        @collapse="onCollapse"
        width="200"
      >
        <div class="logo" style="height: 32px; margin: 16px; background: rgba(255, 255, 255, 0.3); text-align: center; line-height: 32px;">
          后台系统
        </div>
  
        <a-menu
          theme="dark"
          mode="inline"
          :selectedKeys="[activeMenu]"
          :openKeys="openKeys"
          @click="onMenuClick"
          :items="menuItems"
        />
      </a-layout-sider>
  
      <!-- 右侧内容布局 -->
      <a-layout>
        <!-- 顶部 Header -->
        <!-- <LayoutHeader @clearPreferencesAndLogout="logout" /> -->
  
        <!-- 面包屑 -->
        <a-breadcrumb style="margin: 16px">
          <a-breadcrumb-item v-for="(item, index) in breadcrumb" :key="index">
            {{ item }}
          </a-breadcrumb-item>
        </a-breadcrumb>
  
        <!-- 主体内容 -->
        <a-layout-content style="margin: 16px; padding: 24px; background: #fff; min-height: 280px">
          <router-view />
        </a-layout-content>
  
        <!-- 底部 -->
        <a-layout-footer style="text-align: center">
          ©2026 后台管理系统
        </a-layout-footer>
      </a-layout>
    </a-layout>
  </template>
  
  <script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
//   import LayoutHeader from '@/components/LayoutHeader.vue';
//   import { useUserStore } from '@/stores/user';
  import type { MenuProps } from 'ant-design-vue';
  
  // 左侧菜单示例
  const menuItems: MenuProps['items'] = [
    { label: '首页', key: '/dashboard' },
    { label: '用户管理', key: '/users' },
    { label: '系统设置', key: '/settings' },
  ];
  
  // 侧边栏折叠状态
  const collapsed = ref(false);
  const onCollapse = (collapse: boolean) => {
    collapsed.value = collapse;
  };
  const toggleCollapsed = () => {
    collapsed.value = !collapsed.value;
  };
  
  // 路由相关
  const route = useRoute();
  const router = useRouter();
//   const userStore = useUserStore();
  
  const activeMenu = computed(() => route.path);
  const openKeys = ref<string[]>([]);
  
  const onMenuClick = (e: any) => {
    router.push(e.key);
  };
  
  // 面包屑生成
  const breadcrumb = computed(() => {
    const paths = route.path.split('/').filter(Boolean);
    return paths.length ? paths : ['首页'];
  });
  
  // 登出函数
  function logout() {
    // userStore.clearToken();
    router.push('/login');
  }
  </script>
  
  <style scoped>
  .logo {
    color: white;
    font-size: 20px;
    font-weight: bold;
  }
  </style>