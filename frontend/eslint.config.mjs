// @ts-check

import { defineConfig } from '@vben/eslint-config';

export default defineConfig([
  {
    rules: {
      // prettier 对多行元素结束标签输出 `</tag\n>` 格式，与 vue/html-closing-bracket-newline
      // 的 singleline: never 存在固有冲突（eslint --fix 会循环修复），由 prettier 统一格式
      'vue/html-closing-bracket-newline': 'off',
    },
  },
]);
