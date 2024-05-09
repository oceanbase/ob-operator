import type { DrawerProps } from 'antd';
import { Button, Drawer, Space } from 'antd';

import styles from './index.less';

type AlertRuleDrawerProps = {
  onSubmit: () => void;
} & DrawerProps;
export default function AlertDrawer({
  onClose,
  onSubmit,
<<<<<<< HEAD
  footer,
  ...props
}: AlertRuleDrawerProps) {
  return (
    <Drawer closable={false} onClose={onClose} {...props}>
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
=======
  ...props
}: AlertRuleDrawerProps) {
  return (
    <Drawer onClose={onClose} {...props}>
      {props.children}
      <div className={styles.drawerFooter}>
        <Space>
          <Button onClick={onSubmit} type="primary">
            提交
          </Button>
          <Button onClick={onClose}>取消</Button>
        </Space>
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
      </div>
    </Drawer>
  );
}
