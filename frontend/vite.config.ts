import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from 'path'
import tailwindcss from '@tailwindcss/vite'

const pathResolve = (dir: string): any => {
  return path.resolve(__dirname, ".", dir);
};

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": pathResolve("./src/"),
      assets: pathResolve("./src/assets"),
    },
  },
  server: {
    host: true,
    hmr: {
      host: 'localhost',
    },
  },
  // css: {
  //   preprocessorOptions: {
  //     scss: {
  //       // 方法 1: 自动注入变量文件 (推荐用于变量/混入)
  //       // 注意：这里不要写 @import，而是直接写内容，或者使用 javascript API 的 additionalData
  //       additionalData: `@import "${path.resolve(__dirname, 'src/styles/variables.scss')}";`,
        
  //       // 如果只需要导入一次且不想在每个文件重复解析，也可以只放变量定义
  //       // 确保 src/styles/variables.scss 只有变量定义，没有具体的样式输出
  //     }
  //   }
  // }
});
