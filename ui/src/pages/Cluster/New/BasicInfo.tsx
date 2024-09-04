import { intl } from '@/utils/intl';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';

import PasswordInput from '@/components/PasswordInput';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import TooltipPretty from '@/components/TooltipPretty';
import { MODE_MAP } from '@/constants';
import { resourceNameRule } from '@/constants/rules';

interface BasicInfoProps {
  form: FormInstance<API.CreateClusterData>;
  passwordVal: string;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}

export default function BasicInfo({
  form,
  passwordVal,
  setPasswordVal,
}: BasicInfoProps) {
  return (
    <Card
      title={intl.formatMessage({
        id: 'dashboard.Cluster.New.BasicInfo.BasicInformation',
        defaultMessage: '基本信息',
      })}
    >
      <Row gutter={[16, 32]}>
        <Col span={8} style={{ height: 48 }}>
          <SelectNSFromItem form={form} />
        </Col>
        <Col span={8} style={{ height: 48 }}>
          <PasswordInput
            value={passwordVal}
            onChange={setPasswordVal}
            form={form}
            name="rootPassword"
          />
        </Col>
        <Col span={8} style={{ height: 48 }}>
          <TooltipPretty
            title={intl.formatMessage({
              id: 'OBDashboard.Cluster.New.BasicInfo.TheNameOfTheResource',
              defaultMessage: 'k8s中资源的名称',
            })}
          >
            <Form.Item
              label={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.BasicInfo.ResourceName',
                defaultMessage: '资源名称',
              })}
              name="name"
              validateTrigger="onChange"
              validateFirst
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BasicInfo.EnterAKSResource',
                    defaultMessage: '请输入k8s资源名称',
                  }),
                },
                {
                  pattern: /\D/,
                  message: intl.formatMessage({
                    id: 'Dashboard.Cluster.New.BasicInfo.ResourceNamesCannotUsePure',
                    defaultMessage: '资源名不能使用纯数字',
                  }),
                },
                resourceNameRule,
              ]}
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.EnterAResourceName',
                  defaultMessage: '请输入资源名',
                })}
              />
            </Form.Item>
          </TooltipPretty>
        </Col>

        <Col span={8} style={{ height: 72 }}>
          <Form.Item
            label={intl.formatMessage({
              id: 'OBDashboard.Cluster.New.BasicInfo.ClusterName',
              defaultMessage: '集群名',
            })}
            name="clusterName"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.EnterAClusterName',
                  defaultMessage: '请输入集群名',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.BasicInfo.EnterAClusterName',
                defaultMessage: '请输入集群名',
              })}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'Dashboard.Cluster.New.BasicInfo.ClusterMode',
              defaultMessage: '集群模式',
            })}
            name="mode"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Cluster.New.BasicInfo.SelectClusterMode',
                  defaultMessage: '请选择集群模式',
                }),
              },
            ]}
          >
            <Select
              placeholder={intl.formatMessage({
                id: 'Dashboard.Cluster.New.BasicInfo.PleaseSelect',
                defaultMessage: '请选择',
              })}
              optionLabelProp="selectLabel"
              options={Array.from(MODE_MAP.keys()).map((key) => ({
                value: key,
                selectLabel: MODE_MAP.get(key)?.text,
                label: (
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                    }}
                  >
                    <span>{MODE_MAP.get(key)?.text}</span>
                    <span>{MODE_MAP.get(key)?.limit}</span>
                  </div>
                ),
              }))}
            />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  );
}
