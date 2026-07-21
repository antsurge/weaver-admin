import '@vben/request';

declare module 'axios' {
  interface AxiosRequestConfig {
    /**
     * 是否显示成功提示（true 时使用默认文案或后端 message）
     */
    showSuccessMessage?: boolean;
    /**
     * 成功提示文案；为 true 时使用默认文案或后端 message
     */
    successMessage?: boolean | string;
    /**
     * 是否显示失败提示
     */
    showFailMessage?: boolean;
    /**
     * 失败提示文案
     */
    failMessage?: string;
  }
}
