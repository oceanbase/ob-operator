import { intl } from '@/utils/intl';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Checkbox, Col, Input, Popconfirm, Row, Space } from 'antd';

type Label = {
  [T: string]: string | boolean;
};

interface InputLabelCompPorps {
  value?: Label[];
  onChange?: (value: Label[]) => void;
  onBlur?: () => void;
  maxLength?: number;
  defaulLabelName?: string;
  regex?: boolean;
  disable?: boolean;
  allowDelete?: boolean;
}

export default function InputLabelComp(props: InputLabelCompPorps) {
  const {
    value: labels = [],
    onChange,
    maxLength,
    onBlur,
    defaulLabelName = 'key',
    regex,
    disable = false,
    allowDelete = true,
  } = props;

  const labelNameInput = (value: string, index: number) => {
    labels[index][defaulLabelName] = value;
    onChange?.([...labels]);
  };
  const labelValueInput = (value: string, index: number) => {
    labels[index].value = value;
    onChange?.([...labels]);
  };
  const regChange = (checked: boolean, index: number) => {
    labels[index].isRegex = checked;
    onChange?.([...labels]);
  };
  const add = () => {
    const temp: Label = {
      [defaulLabelName]: '',
      value: '',
    };
    if (regex) temp.isRegex = false;
    onChange?.([...labels, temp]);
  };
  const remove = (index: number) => {
    labels.splice(index, 1);
    onChange?.([...labels]);
  };
  return (
    <div>
      <Space style={{ width: '100%', marginBottom: 12 }} direction="vertical">
        {labels?.map((label, index) => (
          <Row gutter={[12, 12]} style={{ alignItems: 'center' }} key={index}>
            <Col span={regex ? 11 : 12}>
              <Input
                disabled={disable}
                value={label[defaulLabelName] as string}
                onBlur={onBlur}
                onChange={(e) => labelNameInput(e.target.value, index)}
                placeholder={intl.formatMessage({
                  id: 'src.components.InputLabelComp.36DDAD02',
                  defaultMessage: '请输入标签名',
                })}
              />
            </Col>
            <Col span={regex ? 10 : 11}>
              <Input
                disabled={disable}
                value={label.value as string}
                onBlur={onBlur}
                onChange={(e) => labelValueInput(e.target.value, index)}
                placeholder={intl.formatMessage({
                  id: 'src.components.InputLabelComp.DC297030',
                  defaultMessage: '请输入标签值',
                })}
              />
            </Col>
            {regex && (
              <Col span={2}>
                <Checkbox
                  checked={label.isRegex as boolean}
                  onChange={(e) => regChange(e.target.checked, index)}
                />

                <span style={{ marginLeft: 8 }}>
                  {intl.formatMessage({
                    id: 'src.components.InputLabelComp.1617B8B8',
                    defaultMessage: '正则',
                  })}
                </span>
              </Col>
            )}

            {labels.length > 1 && allowDelete && (
              <Popconfirm
                placement="left"
                title={intl.formatMessage({
                  id: 'src.components.InputLabelComp.F6592C30',
                  defaultMessage: '确定要删除该配置项吗？',
                })}
                onConfirm={() => {
                  remove(index);
                }}
                okText={intl.formatMessage({
                  id: 'src.components.InputLabelComp.C1E48711',
                  defaultMessage: '删除',
                })}
                cancelText={intl.formatMessage({
                  id: 'src.components.InputLabelComp.D089EBDE',
                  defaultMessage: '取消',
                })}
                okButtonProps={{
                  danger: true,
                  ghost: true,
                }}
              >
                <DeleteOutlined style={{ color: 'rgba(0, 0, 0, .45)' }} />
              </Popconfirm>
            )}
          </Row>
        ))}
      </Space>

      {(!maxLength || labels.length < maxLength) && !disable ? (
        <Row>
          <Col span={24}>
            <Button
              type="dashed"
              block
              onClick={add}
              style={{ color: 'rgba(0,0,0,0.65)' }}
            >
              <PlusOutlined />
              {intl.formatMessage({
                id: 'src.components.InputLabelComp.6C88A39D',
                defaultMessage: '添加',
              })}
            </Button>
          </Col>
        </Row>
      ) : null}
    </div>
  );
}
