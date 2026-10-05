/**
 * 低代码模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const LowcodeAuthCode = {
  /** 低代码（顶级目录） */
  ROOT: 'Lowcode',
  FormBuilder: {
    /** 表单构建器 */
    ROOT: 'Lowcode:FormBuilder',
  },
  Codegen: {
    /** 代码生成器 */
    ROOT: 'Lowcode:Codegen',
  },
} as const;

export type LowcodeAuthCodeValue =
  (typeof LowcodeAuthCode)[keyof typeof LowcodeAuthCode];
