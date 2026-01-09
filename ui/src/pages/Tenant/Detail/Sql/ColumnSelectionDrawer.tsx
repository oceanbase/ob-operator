import { intl } from '@/utils/intl';
import {
  Button,
  Card,
  Checkbox,
  Col,
  Drawer,
  Empty,
  Row,
  Space,
  Typography,
} from 'antd';
import { useEffect, useState } from 'react';

interface ColumnSelectionDrawerProps {
  open: boolean;
  onClose: () => void;
  selectedKeys: string[];
  onSelectionChange: (keys: string[]) => void;
  metrics: API.SqlMetricMetaCategory[];
}

export default function ColumnSelectionDrawer({
  open,
  onClose,
  selectedKeys,
  onSelectionChange,
  metrics,
}: ColumnSelectionDrawerProps) {
  const [currentSelectedKeys, setCurrentSelectedKeys] =
    useState<string[]>(selectedKeys);

  // Sync currentSelectedKeys with parent's selectedKeys whenever drawer opens or parent's selection changes
  useEffect(() => {
    setCurrentSelectedKeys(selectedKeys);
  }, [open, selectedKeys]); // Added open to dependency array to sync when drawer opens

  const handleApply = () => {
    onSelectionChange(currentSelectedKeys);
    onClose();
  };

  const handleCheckboxChange = (key: string, checked: boolean) => {
    if (checked) {
      setCurrentSelectedKeys([...currentSelectedKeys, key]);
    } else {
      setCurrentSelectedKeys(currentSelectedKeys.filter((k) => k !== key));
    }
  };

  return (
    <Drawer
      title={intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.ColumnSelection',
        defaultMessage: '列选择',
      })}
      width={600}
      onClose={onClose}
      open={open}
      footer={
        <Space style={{ float: 'right' }}>
          <Button onClick={onClose}>
            {intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.Cancel',
              defaultMessage: '取消',
            })}
          </Button>
          <Button type="primary" onClick={handleApply}>
            {intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.Ok',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      }
    >
      {metrics && metrics.length > 0 ? ( // Check if metrics array has data
        metrics.map((category) => (
          <Card
            key={category.category}
            title={category.category}
            size="small"
            style={{ marginBottom: 16 }}
          >
            <Row gutter={[16, 8]}>
              {category.metrics.map((metric) => (
                <Col span={12} key={metric.key}>
                  <Checkbox
                    checked={currentSelectedKeys.includes(metric.key)}
                    disabled={metric.immutable}
                    onChange={(e) =>
                      handleCheckboxChange(metric.key, e.target.checked)
                    }
                  >
                    {metric.name}
                  </Checkbox>
                  <Typography.Text
                    type="secondary"
                    style={{ fontSize: 12, display: 'block', marginLeft: 24 }}
                  >
                    {metric.description}
                  </Typography.Text>
                </Col>
              ))}
            </Row>
          </Card>
        ))
      ) : (
        <Empty
          description={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Sql.NoMetricsAvailable',
            defaultMessage: '暂无指标可用',
          })}
        />
      )}
    </Drawer>
  );
}
