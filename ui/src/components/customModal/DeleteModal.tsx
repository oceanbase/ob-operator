import { intl } from '@/utils/intl';
import { ExclamationCircleFilled } from '@ant-design/icons';
import { Modal } from 'antd';

interface showDeleteConfirmProps {
  onOk: () => void;
  onCancel?: () => void;
  title: string;
  content?: string;
}

const { confirm } = Modal;
export default function showDeleteConfirm({
  onOk,
  title,
  content,
  onCancel,
}: showDeleteConfirmProps) {
  confirm({
    title,
    icon: <ExclamationCircleFilled />,
    content,
    okText: intl.formatMessage({
      id: 'OBDashboard.components.customModal.DeleteModal.Yes',
      defaultMessage: '是的',
    }),
    okType: 'danger',
    cancelText: intl.formatMessage({
      id: 'OBDashboard.components.customModal.DeleteModal.Cancel',
      defaultMessage: '取消',
    }),
    onCancel,
    onOk,
  });
}
