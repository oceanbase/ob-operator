import { intl } from '@/utils/intl';
import type { DrawerProps } from 'antd';
import { Button, Drawer, Space } from 'antd';

import styles from './index.less';

type AlertRuleDrawerProps = {
  onSubmit: () => void;
} & DrawerProps;
export default function AlertDrawer({
  onClose,
  onSubmit,
  footer,
  ...props
}: AlertRuleDrawerProps) {
  return (
    <Drawer
      style={{ paddingBottom: 60 }}
      closeIcon={false}
      onClose={onClose}
      maskClosable={false}
      {...props}
    >
      {props.children}
      <div className={styles.drawerFooter}>
        {footer ? (
          footer
        ) : (
          <Space>
            <Button onClick={onSubmit} type="primary">
              {intl.formatMessage({
                id: 'src.components.AlertDrawer.95C6A631',
                defaultMessage: '提交',
              })}
            </Button>
            <Button onClick={onClose}>
              {intl.formatMessage({
                id: 'src.components.AlertDrawer.9B7CD984',
                defaultMessage: '取消',
              })}
            </Button>
          </Space>
        )}
      </div>
    </Drawer>
  );
}
