import IconTip from '@/components/IconTip';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import { getSimpleClusterList } from '@/services';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';

interface BasicConfigProps {
  form: FormInstance<any>;
}

export default function BasicConfig({ form }: BasicConfigProps) {
  const { data: clusterListRes } = useRequest(getSimpleClusterList);
  const clisterList = clusterListRes?.data.map((cluster) => ({
    label: cluster.name,
    value: `${cluster.name}+${cluster.namespace}`,
  }));
  return (
    <Card
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.New.F830C6B7',
        defaultMessage: '基本设置',
      })}
    >
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.D41C48E1',
              defaultMessage: '资源名称',
            })}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.OBProxy.New.A0CBFC04',
                  defaultMessage: '请输入',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.E9A192AB',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.2D601471',
              defaultMessage: 'OBProxy 集群名',
            })}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.OBProxy.New.6B7B9E9A',
                  defaultMessage: '请输入',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.DEF127D3',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.7D1609DD',
              defaultMessage: '连接 OB 集群',
            })}
            name="obCluster"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.OBProxy.New.80C781AA',
                  defaultMessage: '请选择',
                }),
              },
            ]}
          >
            <Select options={clisterList} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={
              <IconTip
                content={intl.formatMessage({
                  id: 'src.pages.OBProxy.New.E711F60D',
                  defaultMessage: 'OBProxy root 密码',
                })}
                tip={intl.formatMessage({
                  id: 'src.pages.OBProxy.New.26311C9E',
                  defaultMessage: 'root@proxysys 密码',
                })}
              />
            }
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.OBProxy.New.D76A01D7',
                  defaultMessage: '请输入',
                }),
              },
            ]}
          >
            <Input.Password
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.B2851499',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <SelectNSFromItem form={form} />
        </Col>
      </Row>
    </Card>
  );
}
