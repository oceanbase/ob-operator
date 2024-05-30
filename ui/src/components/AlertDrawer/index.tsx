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
      closable={false}
      onClose={onClose}
      {...props}
    >
      {props.children}
      <div className={styles.drawerFooter}>
        {footer ? (
          footer
        ) : (
          <Space>
            <Button onClick={onSubmit} type="primary">
              提交
            </Button>
            <Button onClick={onClose}>取消</Button>
          </Space>
        )}
      </div>
    </Drawer>
  );
}
