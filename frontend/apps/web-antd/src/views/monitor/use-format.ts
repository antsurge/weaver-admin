/** 字节数格式化为可读字符串 */
export function formatBytes(bytes?: number): string {
  const value = Number(bytes ?? 0);
  if (!Number.isFinite(value) || value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = value;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(2)} ${units[i]}`;
}

/** 秒数格式化为「x天x小时x分x秒」 */
export function formatDuration(seconds?: number): string {
  const total = Math.max(0, Math.floor(Number(seconds ?? 0)));
  const d = Math.floor(total / 86_400);
  const h = Math.floor((total % 86_400) / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const parts: string[] = [];
  if (d > 0) parts.push(`${d}天`);
  if (h > 0) parts.push(`${h}小时`);
  if (m > 0) parts.push(`${m}分`);
  parts.push(`${s}秒`);
  return parts.join('');
}

/** TTL 展示：-1 表示永不过期 */
export function formatTtl(seconds?: number): string {
  const v = Number(seconds ?? -1);
  if (v < 0) return '永不过期';
  return formatDuration(v);
}
