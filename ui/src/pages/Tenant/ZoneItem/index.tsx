import InputNumber from '@/components/InputNumber';
import { intl } from '@/utils/intl';
import { Checkbox, Col, Form } from 'antd';

interface ZoneItemProps {
  name: string;
  checked: boolean;
  obZoneResource: any;
  checkBoxOnChange: (checked: boolean, name: string) => void;
  key: number;
  formName?:string[]|string;
  isEdit?:boolean;
}

export default function ZoneItem({
  name,
  key,
  checked,
  obZoneResource,
  checkBoxOnChange,
  isEdit,
  formName = ['pools', name, 'priority']
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
      <Checkbox
        checked={checked}
        disabled={isEdit}
        style={{ marginRight: 24 }}
        onChange={(e) => checkBoxOnChange(e.target.checked, name)}
      />

      <Col span={4}>
        <Form.Item
          name={formName}
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
