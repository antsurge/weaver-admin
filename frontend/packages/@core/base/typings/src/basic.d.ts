import { PermissionMenuApi } from '@/api/permission/menu';

interface BasicOption {
  label: string;
  value: string;
}

type SelectOption = BasicOption;

type TabOption = BasicOption;

interface BasicUserInfo {
  /**
   * 头像
   */
  avatar: string;
  /**
   * 归属部门名称（个人中心只读展示）
   */
  departmentName?: string;
  /**
   * 菜单树
   */
  menuTree: PermissionMenuApi.PermissionMenu[];
  /**
   * 用户昵称
   */
  realName: string;
  /**
   * 角色code
   */
  roleCodes: string[];
  /**
   * 角色名称列表（个人中心只读展示）
   */
  roleNames?: string[];
  /**
   * 用户角色
   */
  roles?: string[];
  /**
   * 用户id
   */
  userId: string;
  /**
   * 用户名
   */
  username: string;
}

type ClassType = Array<object | string> | object | string;

export type { BasicOption, BasicUserInfo, ClassType, SelectOption, TabOption };
