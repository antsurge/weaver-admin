import { message } from 'ant-design-vue';

export async function downloadFile(data: BlobPart, fileName: string) {
  const blob = new Blob([data]);

  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = fileName;

  document.body.append(link);
  link.click();

  link.remove();
  URL.revokeObjectURL(link.href);
}

export function getFileNameFromDisposition(disposition?: string) {
  if (!disposition) return '';

  // ⭐ 优先匹配 filename*
  const filenameStarMatch = disposition.match(/filename\*\s*=\s*UTF-8''(.+)/i);
  if (filenameStarMatch?.[1]) {
    return decodeURIComponent(filenameStarMatch[1]);
  }

  // 兜底 filename
  const filenameMatch = disposition.match(/filename="?([^"]+)"?/);
  if (filenameMatch?.[1]) {
    return filenameMatch[1];
  }

  return '';
}

/**
 * 处理 blob 请求（如导出）失败时的错误提示。
 * 后端出错时会返回 JSON 错误体，但响应被当作 blob 读取，
 * 这里需要把 blob 还原为文本并解析出 message。
 * 对非 JSON / 无法读取的情况做兜底，避免错误被静默吞掉。
 */
export async function handleBlobResponseError(error: any) {
  let text = '';

  try {
    // error 可能是 Response/Blob（有 text()），也可能是普通 Error/字符串
    if (typeof error?.text === 'function') {
      text = await error.text();
    } else if (error instanceof Blob) {
      text = await error.text();
    } else if (typeof error === 'string') {
      text = error;
    } else {
      text = error?.message ?? '';
    }
  } catch {
    text = '';
  }

  // 尝试解析后端返回的 JSON 错误体，取其中的 message / msg
  if (text) {
    try {
      const json = JSON.parse(text);
      const msg = json?.message ?? json?.msg;
      if (msg) {
        message.error(msg);
        return;
      }
    } catch {
      // 非 JSON（如 HTML 错误页 / 纯文本），走下方兜底
    }
  }

  message.error(text || error?.message || '导出失败，请稍后重试');
}
