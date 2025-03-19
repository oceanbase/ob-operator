import InputNumber from '@/components/InputNumber';
import { intl } from '@/utils/intl';
import { isGte4_2, isGte4_3_3 } from '@/utils/package';
import { Checkbox, Col, Form, Select } from 'antd';

interface ZoneItemProps {
  name: string;
  type: string;
  obversion: string;
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
  type,
  priorityName = ['pools', name, 'priority'],
  checkedFormName,
  obversion,
}: ZoneItemProps) {
  const REPLICA_TYPE_LIST = [
    { value: 'Full', label: '全能型副本' },
    ...(isGte4_2(obversion)
      ? [{ value: 'Readonly', label: '只读型副本' }]
      : []),
    ...(isGte4_3_3(obversion)
      ? [{ value: 'Column', label: '只读列存型副本' }]
      : []),
  ];

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
      <Col
        span={type === 'new' ? 2 : 4}
        style={type === 'tenantBackup' ? { marginTop: 24 } : {}}
      >
        <Form.Item
          name={priorityName}
          label={intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.Priority',
            defaultMessage: '优先级',
          })}
        >
          <InputNumber
            style={type === 'new' ? {} : { width: '90%' }}
            disabled={!checked}
          />
        </Form.Item>
      </Col>
      <Col span={5} style={type === 'tenantBackup' ? { marginTop: 24 } : {}}>
        <Form.Item name={['pools', name, 'type']} label={'副本类型'}>
          <Select options={REPLICA_TYPE_LIST} defaultValue={'Full'} />
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
