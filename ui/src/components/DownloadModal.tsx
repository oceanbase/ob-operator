import { attachment, job } from '@/api';
import { download } from '@/utils/export';
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
  const [showErrorDetails, setShowErrorDetails] = useState(false);

  // 文件下载 - 后端直接返回文件内容
  const { run: downloadAttachment, loading: downloadLoading } = useRequest(
    attachment.downloadAttachment,
    {
      manual: true,
      onSuccess: (data) => {
        if (data) {
          download(data, attachmentValue);
          message.success('文件下载成功');
        } else {
          message.error('下载数据为空');
        }
        onOk();
      },
      onError: () => {
        message.error('下载失败，请重试');
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
                  downloadAttachment(attachmentValue);
                }}
              >
                下载链接
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
      title={`${title}${
        findByValue(modalContent, diagnoseStatus).label || '中...'
      }`}
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
