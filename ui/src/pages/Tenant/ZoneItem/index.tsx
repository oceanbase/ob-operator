import InputNumber from '@/components/InputNumber';
import { intl } from '@/utils/intl';
import { isGte4_2, isGte4_3_3 } from '@/utils/package';
import { Checkbox, Col, Form, Select } from 'antd';

interface ZoneItemProps {
  name: string;
  type: string;
  obVersion: string;
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
  obVersion,
}: ZoneItemProps) {
  const REPLICA_TYPE_LIST = [
    {
      value: 'Full',
      label: intl.formatMessage({
        id: 'src.pages.Tenant.ZoneItem.14D49F75',
        defaultMessage: '全能型副本',
      }),
    },
    ...(isGte4_2(obVersion)
      ? [
          {
            value: 'Readonly',
            label: intl.formatMessage({
              id: 'src.pages.Tenant.ZoneItem.BFD370B4',
              defaultMessage: '只读型副本',
            }),
          },
        ]
      : []),
    ...(isGte4_3_3(obVersion)
      ? [
          {
            value: 'Columnstore',
            label: intl.formatMessage({
              id: 'src.pages.Tenant.ZoneItem.577FB981',
              defaultMessage: '只读列存型副本',
            }),
          },
        ]
      : []),
  ];
  console.log('obZoneResource', obZoneResource);
  return (
    <div
      style={{
        width: '100%',
        display: 'flex',
        justifyContent: 'flex-start',
        alignItems: 'center',
      }}
    >
      {checkedFormName ? (
        <Form.Item noStyle name={checkedFormName}>
          <Checkbox
            checked={checked}
            disabled={isEdit || !obZoneResource}
            style={{ marginRight: 5 }}
            onChange={(e) => checkBoxOnChange(e.target.checked, name)}
          />
        </Form.Item>
      ) : (
        <Checkbox
          checked={checked}
          disabled={isEdit || !obZoneResource}
          style={{ marginRight: 5 }}
          onChange={(e) => checkBoxOnChange(e.target.checked, name)}
        />
      )}
      <span style={{ marginRight: 8 }}>{name}</span>
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
          initialValue={1}
        >
          <InputNumber style={type === 'new' ? {} : { width: '90%' }} min={1} />
        </Form.Item>
      </Col>
      <Col span={5} style={type === 'tenantBackup' ? { marginTop: 24 } : {}}>
        <Form.Item
          name={['pools', name, 'type']}
          label={intl.formatMessage({
            id: 'src.pages.Tenant.ZoneItem.93E193BC',
            defaultMessage: '副本类型',
          })}
          initialValue={isEdit ? obZoneResource['type'] : 'Full'}
        >
          <Select options={REPLICA_TYPE_LIST} />
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
