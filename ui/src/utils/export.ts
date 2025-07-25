// 下载文件和HTML
export const download = (content: string, fileName?: string) => {
  const blob = new Blob([content]);
  const blobUrl = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  if (fileName) {
    a.download = fileName;
  }
  a.href = blobUrl;
  a.click();
  a.remove();
  window.URL.revokeObjectURL(blobUrl);
};

// 直接触发浏览器下载 - 适用于后端已设置正确HTTP头的接口
export const downloadFile = (attachmentId: string, fileName?: string) => {
  // 构建下载URL
  const baseUrl = window.location.origin;
  const downloadUrl = `${baseUrl}/api/v1/attachments/${attachmentId}`;

  // 创建下载链接
  const a = document.createElement('a');
  a.href = downloadUrl;
  if (fileName) {
    a.download = fileName;
  }

  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
};
