import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Checkbox, Col, Form, Input, Popconfirm, Row } from 'antd';
<<<<<<< HEAD
import styles from './index.less'
=======
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)

interface InputLabelProps {
  wrapFormName: string;
  labelFormName: string;
  valueFormName: string;
<<<<<<< HEAD
  regBoxFormName?: string;
  showDelete?: boolean;
=======
  showRegBox?: boolean;
  showDelIcon?: boolean;
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
}

export default function InputLabel({
  wrapFormName,
  labelFormName,
  valueFormName,
<<<<<<< HEAD
  regBoxFormName,
  showDelete = true,
=======
  showRegBox = false,
  showDelIcon = true,
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
}: InputLabelProps) {
  return (
    <Form.List name={wrapFormName}>
      {(fields, { add, remove }) => {
        return (
          <div>
            {fields.map(({ key, name }, index) => (
              <Row gutter={8} style={{ marginBottom: 8 }} key={key}>
                <Col span={12}>
                  <Form.Item name={[name, labelFormName]} noStyle>
                    <Input placeholder="请输入标签名" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={[name, valueFormName]} noStyle>
                    <Input placeholder="请输入标签值" />
                  </Form.Item>
                </Col>
<<<<<<< HEAD
                {regBoxFormName && (
                  <Col span={2}>
                    <Form.Item name={[name, regBoxFormName]} noStyle>
                      <Checkbox style={{ marginRight: 4 }} />
=======
                {showRegBox && (
                  <Col span={4}>
                    <Form.Item name={[name, 'isRegex']} noStyle>
                      <Checkbox />
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
                      正则
                    </Form.Item>
                  </Col>
                )}
<<<<<<< HEAD
                {showDelete && fields.length > 1 && (
                  <Col span={2}>
                    <Form.Item className={styles.delContent} name={[name, ' ']}>
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
=======
                {showDelIcon && fields.length > 1 && (
                  <Form.Item
                    label={index === 0 && ' '}
                    style={{ marginBottom: 8 }}
                    name={[name, ' ']}
                  >
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
                      <DeleteOutlined style={{ color: 'rgba(0, 0, 0, .45)' }} />
                    </Popconfirm>
                  </Form.Item>
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
                )}
              </Row>
            ))}
            <Row>
              <Col span={20}>
                <Form.Item>
                  <Button
                    type="dashed"
                    block
                    onClick={() => {
                      const temp = { labelFormName: '', valueFormName: '' };
<<<<<<< HEAD
                      if (regBoxFormName) temp[regBoxFormName] = false;
=======
                      if (showRegBox) temp.isRegex = false;
>>>>>>> b1fb5a2... Prepare for 2.2.1 test (#357)
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
          </div>
        );
      }}
    </Form.List>
  );
}
