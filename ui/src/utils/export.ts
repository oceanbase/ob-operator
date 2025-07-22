// 下载文件和HTML
export const download = (content: string, fileName?: string) => {
  const blob = new Blob([content]);
  const blobUrl = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.download = fileName;
  a.href = blobUrl;
  a.click();
  a.remove();
  window.URL.revokeObjectURL(blobUrl);
};
