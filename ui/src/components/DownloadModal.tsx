import { attachment, job } from '@/api';
import { LoadingOutlined } from '@ant-design/icons';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Alert, Button, Modal, Spin, message } from 'antd';
import { useState } from 'react';

interface DownloadModalProps {
  visible: boolean;
  onCancel: () => void;
  onOk: () => void;
  title: string;
  diagnoseStatus: string;
  attachmentValue: string;
  jobValue?: any;
  errorLogs?: any;
}

export default function DownloadModal({
  visible,
  onCancel,
  onOk,
  title,
  diagnoseStatus,
  attachmentValue,
  jobValue,
  errorLogs,
}: DownloadModalProps) {
  const [retryCount, setRetryCount] = useState(0);
  const [downloadTimeout, setDownloadTimeout] = useState<NodeJS.Timeout | null>(
    null,
  );
  const [showErrorDetails, setShowErrorDetails] = useState(false);

  // 集群日志下载
  const { run: downloadAttachment, loading: downloadLoading } = useRequest(
    attachment.downloadAttachment,
    {
      manual: true,
      onSuccess: (data) => {
        // 处理zip文件下载
        if (data) {
          try {
            // 检查数据大小
            if (data.size === 0) {
              message.error('下载的文件为空，请重试');
              return;
            }

            // 创建Blob对象
            const blob = new Blob([data], { type: 'application/zip' });

            // 检查Blob大小
            if (blob.size === 0) {
              message.error('文件内容为空，请重试');
              return;
            }

            // 创建下载链接
            const url = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;

            // 设置文件名
            const fileName = `cluster_log_${new Date().getTime()}.zip`;
            link.download = fileName;

            // 触发下载
            document.body.appendChild(link);
            link.click();

            // 清理
            document.body.removeChild(link);
            window.URL.revokeObjectURL(url);

            // 显示成功消息
            message.success('文件下载成功');
            // 清除超时和重试计数
            if (downloadTimeout) {
              clearTimeout(downloadTimeout);
              setDownloadTimeout(null);
            }
            setRetryCount(0);
          } catch (error) {
            message.error('下载文件失败，请重试');
            // 增加重试计数
            setRetryCount((prev) => prev + 1);
          }
        } else {
          message.error('下载数据为空');
          // 增加重试计数
          setRetryCount((prev) => prev + 1);
        }
        onOk();
      },
      onError: () => {
        message.error('下载失败，请重试');
        // 增加重试计数
        setRetryCount((prev) => prev + 1);
        // 清除超时
        if (downloadTimeout) {
          clearTimeout(downloadTimeout);
          setDownloadTimeout(null);
        }
      },
    },
  );

  const { run: deleteJob } = useRequest(job.deleteJob, {
    manual: true,
    onSuccess: () => {
      onCancel();
    },
  });

  const modalContent = [
    {
      value: 'pending',
      label: '中...',
      children: (
        <Spin
          indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />}
          tip="Loading"
        >
          <div style={{ minHeight: 100 }} />
        </Spin>
      ),
    },
    {
      value: 'running',
      label: '中...',
      children: (
        <Spin
          indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />}
          tip="Loading"
        >
          <div style={{ minHeight: 100 }} />
        </Spin>
      ),
    },
    {
      value: 'successful',
      label: '完成',
      children: (
        <div style={{ minHeight: 100 }}>
          <Alert
            message={'信息收集与分析完成'}
            type="success"
            showIcon
            action={
              <Button
                size="small"
                type="link"
                loading={downloadLoading}
                onClick={() => {
                  // 清除之前的超时
                  if (downloadTimeout) {
                    clearTimeout(downloadTimeout);
                  }

                  // 设置下载超时（5分钟）
                  const timeout = setTimeout(() => {
                    message.error('下载超时，请重试');
                    setRetryCount(0);
                  }, 5 * 60 * 1000);
                  setDownloadTimeout(timeout);

                  downloadAttachment(attachmentValue);
                }}
              >
                {retryCount > 0 ? `重试下载 (${retryCount})` : '下载链接'}
              </Button>
            }
          />
        </div>
      ),
    },
    {
      value: 'failed',
      label: '失败',
      children: (
        <div style={{ minHeight: 100 }}>
          <Alert
            message={`${title}分析失败`}
            type="error"
            showIcon
            action={
              <Button
                size="small"
                type="link"
                onClick={() => {
                  setShowErrorDetails(true);
                }}
              >
                查看详情
              </Button>
            }
          />

          {showErrorDetails && (
            <div style={{ marginTop: 16 }}>
              <div
                style={{
                  backgroundColor: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  border: '1px solid #d9d9d9',
                  maxHeight: 200,
                  overflow: 'auto',
                  fontFamily: 'monospace',
                  fontSize: 12,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                }}
              >
                <div
                  style={{
                    fontWeight: 'bold',
                    marginBottom: 8,
                    color: '#ff4d4f',
                  }}
                >
                  错误详情：
                </div>
                {errorLogs}
              </div>
              <Button
                size="small"
                type="link"
                style={{ marginTop: 8 }}
                onClick={() => setShowErrorDetails(false)}
              >
                隐藏详情
              </Button>
            </div>
          )}
        </div>
      ),
    },
  ];

  return (
    <Modal
      title={`${title}${findByValue(modalContent, diagnoseStatus).label}`}
      open={visible}
      onOk={onOk}
      onCancel={() => deleteJob(jobValue?.namespace, jobValue?.name)}
      footer={
        diagnoseStatus === 'running' ? (
          <Button
            onClick={() => deleteJob(jobValue?.namespace, jobValue?.name)}
          >
            取消
          </Button>
        ) : null
      }
    >
      {findByValue(modalContent, diagnoseStatus)?.children}
    </Modal>
  );
}
