import { intl } from '@/utils/intl';
import { ExclamationCircleFilled } from '@ant-design/icons';
import type { ModalFuncProps } from 'antd';
import { Modal } from 'antd';
import { debounce } from 'lodash';
import styles from './index.less';

const { confirm } = Modal;
function showDeleteConfirm(props: ModalFuncProps) {
  confirm({
    icon: <ExclamationCircleFilled />,
    okText: intl.formatMessage({
      id: 'OBDashboard.components.customModal.DeleteModal.Yes',
      defaultMessage: '是的',
    }),
    okType: 'danger',
    cancelText: intl.formatMessage({
      id: 'OBDashboard.components.customModal.DeleteModal.Cancel',
      defaultMessage: '取消',
    }),
    className: styles.deleteContainer,
    ...props,
  });
}
export default debounce(showDeleteConfirm, 500, {
  leading: true,
  trailing: false,
});
