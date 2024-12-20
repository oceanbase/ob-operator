import InputNumber from '@/components/InputNumber';
import { intl } from '@/utils/intl';
import { Checkbox, Col, Form } from 'antd';

interface ZoneItemProps {
  name: string;
  checked: boolean;
  obZoneResource: API.ZoneResource;
  checkBoxOnChange: (checked: boolean, name: string) => void;
  priorityName?: string[] | string;
  isEdit?: boolean;
  checkedFormName?: string[] | string;
}

export default function ZoneItem({
  name,
  checked,
  obZoneResource,
  checkBoxOnChange,
  isEdit,
  priorityName = ['pools', name, 'priority'],
  checkedFormName,
}: ZoneItemProps) {
  return (
    <div
      style={{
        width: '100%',
        display: 'flex',
        justifyContent: 'flex-start',
        alignItems: 'center',
      }}
    >
      <span style={{ marginRight: 8 }}>{name}</span>
      {checkedFormName ? (
        <Form.Item noStyle name={checkedFormName}>
          <Checkbox
            checked={checked}
            disabled={isEdit || !obZoneResource}
            style={{ marginRight: 24 }}
            onChange={(e) => checkBoxOnChange(e.target.checked, name)}
          />
        </Form.Item>
      ) : (
        <Checkbox
          checked={checked}
          disabled={isEdit || !obZoneResource}
          style={{ marginRight: 24 }}
          onChange={(e) => checkBoxOnChange(e.target.checked, name)}
        />
      )}
      <Col span={4}>
        <Form.Item
          name={priorityName}
          label={intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.Priority',
            defaultMessage: '优先级',
          })}
        >
          <InputNumber style={{ width: '100%' }} disabled={!checked} />
        </Form.Item>
      </Col>
      {obZoneResource && (
        <Col style={{ marginLeft: 12 }} span={16}>
          <span style={{ marginRight: 12 }}>
            {intl.formatMessage({
              id: 'Dashboard.Tenant.New.ResourcePools.AvailableResources',
              defaultMessage: '可用资源：',
            })}
          </span>
          <span style={{ marginRight: 12 }}>
            CPU {obZoneResource['availableCPU']}
          </span>
          <span style={{ marginRight: 12 }}>
            Memory {obZoneResource['availableMemory']}GB
          </span>
          <span>
            {intl.formatMessage({
              id: 'Dashboard.Tenant.New.ResourcePools.LogDiskSize',
              defaultMessage: '日志磁盘大小',
            })}
            {obZoneResource['availableLogDisk']}GB
          </span>
        </Col>
      )}
    </div>
  );
}
