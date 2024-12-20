import { intl } from '@/utils/intl';
import { Card, Checkbox, Col, Form, Input, Row, Select, Space } from 'antd';
import type { FormInstance } from 'antd/lib/form';

import PasswordInput from '@/components/PasswordInput';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import TooltipPretty from '@/components/TooltipPretty';
import { LOADTYPE_LIST, MODE_MAP } from '@/constants';
import { resourceNameRule } from '@/constants/rules';

const { Option } = Select;

interface BasicInfoProps {
  form: FormInstance<API.CreateClusterData>;
  passwordVal: string;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}

export default function BasicInfo({
  form,
  passwordVal,
  deleteValue,
  setPasswordVal,
  setDeleteValue,
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
        <Col span={8}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.Cluster.New.04047AB8',
              defaultMessage: '优化场景',
            })}
            name={'scenario'}
            initialValue="HTAP"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.Cluster.New.8126F0B5',
                  defaultMessage: '请选择优化场景',
                }),
              },
            ]}
          >
            <Select>
              {LOADTYPE_LIST?.map((item) => (
                <Option key={item.value} value={item.value}>
                  {item.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>
      <Space>
        {intl.formatMessage({
          id: 'src.pages.Cluster.New.FCB7C4F3',
          defaultMessage: '删除保护',
        })}

        <Checkbox
          defaultChecked={deleteValue}
          onChange={(e) => {
            setDeleteValue(e.target.value);
          }}
        />
      </Space>
    </Card>
  );
}
