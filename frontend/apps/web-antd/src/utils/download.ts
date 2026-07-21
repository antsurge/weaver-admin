import { message } from "ant-design-vue";

export async function downloadFile(data: BlobPart, fileName: string) {
  const blob = new Blob([data]);

  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = fileName;

  document.body.appendChild(link);
  link.click();

  document.body.removeChild(link);
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

export async function handleBlobResponseError(error:any) {
  const text = await error.text()
  const jsonText = JSON.parse(text)
  message.error(jsonText.message)
}
