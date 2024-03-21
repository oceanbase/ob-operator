import InputNumber from '@/components/InputNumber';
import { intl } from '@/utils/intl';
import { Checkbox,Col,Form } from 'antd';

interface ZoneItemProps {
  name: string;
  checked: boolean;
  obZoneResource: any;
  checkBoxOnChange: (checked: boolean, name: string) => void;
  key: number;
  priorityName?:string[]|string;
  isEdit?:boolean;
  checkedFormName?:string[]|string;
}

export default function ZoneItem({
  name,
  key,
  checked,
  obZoneResource,
  checkBoxOnChange,
  isEdit,
  priorityName = ['pools', name, 'priority'],
  checkedFormName
}: ZoneItemProps) {
  return (
    <div
      key={key}
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
            disabled={isEdit}
            style={{ marginRight: 24 }}
            onChange={(e) => checkBoxOnChange(e.target.checked, name)}
          />
        </Form.Item>
      ) : (
        <Checkbox
          checked={checked}
          disabled={isEdit}
          style={{ marginRight: 24 }}
          onChange={(e) => checkBoxOnChange(e.target.checked, name)}
        />
      )}
      <Col span={4}>
        <Form.Item
          name={priorityName}
          label={intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.Weight',
            defaultMessage: '权重',
          })}
        >
          <InputNumber style={{ width: '100%' }} disabled={!checked} />
        </Form.Item>
      </Col>
      {obZoneResource && (
        <Col style={{ marginLeft: 12 }} span={12}>
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
            Memory {obZoneResource['availableMemory'] / (1 << 30)}GB
          </span>
          <span>
            {intl.formatMessage({
              id: 'Dashboard.Tenant.New.ResourcePools.LogDiskSize',
              defaultMessage: '日志磁盘大小',
            })}
            {obZoneResource['availableLogDisk'] / (1 << 30)}GB
          </span>
        </Col>
      )}
    </div>
  );
}
