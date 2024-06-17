import IconTip from '@/components/IconTip';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import TooltipPretty from '@/components/TooltipPretty';
import { resourceNameRule } from '@/constants/rules';
import { getSimpleClusterList } from '@/services';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';
import { useEffect } from 'react';

interface BasicConfigProps {
  form: FormInstance<any>;
}

export default function BasicConfig({ form }: BasicConfigProps) {
  const { data: clusterListRes } = useRequest(getSimpleClusterList);
  const selectCluster = Form.useWatch('obCluster');
  const clisterList = clusterListRes?.data.map((cluster) => ({
    label: cluster.name,
    value: JSON.stringify({ name: cluster.name, namespace: cluster.namespace }),
  }));

  useEffect(() => {
    if (selectCluster && !form.getFieldValue('namespace')) {
      form.setFieldValue('namespace', selectCluster.split('+')?.[1]);
    }
  }, [selectCluster]);
  return (
    <Card
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.New.F830C6B7',
        defaultMessage: '基本设置',
      })}
    >
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <TooltipPretty
            title={intl.formatMessage({
              id: 'src.pages.OBProxy.New.D6D90ACC',
              defaultMessage: 'k8s中资源的名称',
            })}
          >
            <Form.Item
              label={intl.formatMessage({
                id: 'src.pages.OBProxy.New.803427AF',
                defaultMessage: '资源名称',
              })}
              validateFirst
              name={'name'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.OBProxy.New.F602E292',
                    defaultMessage: '请输入k8s资源名称',
                  }),
                },
                {
                  pattern: /\D/,
                  message: intl.formatMessage({
                    id: 'src.pages.OBProxy.New.37FA27BA',
                    defaultMessage: '资源名不能使用纯数字',
                  }),
                },
                resourceNameRule,
              ]}
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'src.pages.OBProxy.New.9B4BA02B',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </TooltipPretty>
        </Col>
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.BB6BC872',
              defaultMessage: 'OBProxy 集群名',
            })}
            name={'proxyClusterName'}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.OBProxy.New.CA42FD5D',
                  defaultMessage: '请输入集群名',
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
                  id: 'src.pages.OBProxy.New.94339826',
                  defaultMessage: '请选择 OB 集群',
                }),
              },
            ]}
          >
            <Select options={clisterList} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={'proxySysPassword'}
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
                  id: 'src.pages.OBProxy.New.67DC144A',
                  defaultMessage: '请输入 OBProxy root 密码',
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
