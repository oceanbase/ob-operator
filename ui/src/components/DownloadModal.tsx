import { attachment, job } from '@/api';
import { LoadingOutlined } from '@ant-design/icons';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Alert, Button, Modal, Spin } from 'antd';

interface DownloadModalProps {
  visible: boolean;
  onCancel: () => void;
  onOk: () => void;
  title: string;
  diagnoseStatus: string;
  attachmentValue: string;
}

export default function DownloadModal({
  visible,
  onCancel,
  onOk,
  title,
  diagnoseStatus,
  attachmentValue,
  jobValue,
}: DownloadModalProps) {
  const { run: downloadAttachment } = useRequest(
    attachment.downloadAttachment,
    {
      manual: true,
      onSuccess: (data) => {
        console.log('data', data);
        onOk();
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
              <Button size="small" type="link">
                查看详情
              </Button>
            }
          />
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
