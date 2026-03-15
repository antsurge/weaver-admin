// 防重复请求
const pendingMap = new Map<string, AbortController>();

function getPendingKey(config: any) {
  const { url, method, params, data } = config;

  return [method, url, JSON.stringify(params), JSON.stringify(data)].join("&");
}

export function addPending(config: any) {
  const key = getPendingKey(config);

  if (!pendingMap.has(key)) {
    const controller = new AbortController();
    config.signal = controller.signal;
    pendingMap.set(key, controller);
  }
}

export function removePending(config: any) {
  const key = getPendingKey(config);

  if (pendingMap.has(key)) {
    const controller = pendingMap.get(key);

    controller?.abort();
    pendingMap.delete(key);
  }
}

export function clearPending() {
  pendingMap.forEach((controller) => controller.abort());
  pendingMap.clear();
}
