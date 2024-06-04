import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import {
  Button,
  Checkbox,
  Col,
  Form,
  FormInstance,
  Input,
  Popconfirm,
  Row,
} from 'antd';
import { ReactElement } from 'react';
import styles from './index.less';

interface InputLabelProps {
  wrapFormName: string;
  labelFormName: string;
  valueFormName: string;
  form: FormInstance<unknown>;
  regBoxFormName?: string;
  showDelete?: boolean;
  maxCount?: number;
}

interface LableFormItemProps {
  name: (string | number)[];
  fieldIdx: number;
  dependName: (string | number)[];
  children: ReactElement;
}

export default function InputLabel({
  wrapFormName,
  labelFormName,
  valueFormName,
  regBoxFormName,
  showDelete = true,
  maxCount,
  form,
}: InputLabelProps) {
  const LableFormItem = ({
    name,
    fieldIdx,
    dependName,
    children,
  }: LableFormItemProps) => {
    return (
      <Form.Item
        dependencies={[
          [wrapFormName, ...dependName],
          [wrapFormName, fieldIdx, regBoxFormName],
        ]}
        noStyle
      >
        {({ getFieldValue, getFieldInstance }) => {
          const rules: unknown[] = [];
          const labelValue = getFieldValue([wrapFormName, ...dependName]);
          const regBox = getFieldInstance([
            wrapFormName,
            fieldIdx,
            regBoxFormName,
          ]);
          if (labelValue || regBox?.input?.checked)
            rules.push({ required: true, message: '请输入' });
          else {
            rules.pop();
            form.validateFields([[wrapFormName, ...name]]);
          }

          return (
            <Form.Item rules={rules} name={name}>
              {children}
            </Form.Item>
          );
        }}
      </Form.Item>
    );
  };
  return (
    <Form.List name={wrapFormName}>
      {(fields, { add, remove }) => {
        return (
          <div>
            {fields.map(({ key, name }, index) => (
              <Row gutter={8} style={{ marginBottom: 8 }} key={key}>
                <Col span={11}>
                  <LableFormItem
                    name={[name, labelFormName]}
                    fieldIdx={name}
                    dependName={[name, valueFormName]}
                  >
                    <Input placeholder="请输入标签名" />
                  </LableFormItem>
                </Col>
                <Col span={10}>
                  <LableFormItem
                    name={[name, valueFormName]}
                    fieldIdx={name}
                    dependName={[name, labelFormName]}
                  >
                    <Input placeholder="请输入标签值" />
                  </LableFormItem>
                </Col>
                {regBoxFormName && (
                  <Col span={2}>
                    <Form.Item name={[name, regBoxFormName]} noStyle>
                      <Checkbox
                        onChange={(e) => {
                          form.setFieldValue(
                            [wrapFormName, name, regBoxFormName],
                            e.target.checked,
                          );
                        }}
                        style={{ marginRight: 4 }}
                      />
                    </Form.Item>
                    正则
                  </Col>
                )}
                {showDelete && fields.length > 1 && (
                  <Col span={1}>
                    <Form.Item className={styles.delContent}>
                      <Popconfirm
                        placement="left"
                        title="确定要删除该配置项吗？"
                        onConfirm={() => {
                          remove(index);
                        }}
                        okText="删除"
                        cancelText="取消"
                        okButtonProps={{
                          danger: true,
                          ghost: true,
                        }}
                      >
                        <DeleteOutlined
                          style={{ color: 'rgba(0, 0, 0, .45)' }}
                        />
                      </Popconfirm>
                    </Form.Item>
                  </Col>
                )}
              </Row>
            ))}
            {!maxCount || fields.length < maxCount ? (
              <Row>
                <Col span={24}>
                  <Form.Item>
                    <Button
                      type="dashed"
                      block
                      onClick={() => {
                        const temp = {
                          [labelFormName]: '',
                          [valueFormName]: '',
                        };
                        if (regBoxFormName) temp[regBoxFormName] = false;
                        add(temp);
                      }}
                      style={{ color: 'rgba(0,0,0,0.65)' }}
                    >
                      <PlusOutlined />
                      添加
                    </Button>
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
          </div>
        );
      }}
    </Form.List>
  );
}
