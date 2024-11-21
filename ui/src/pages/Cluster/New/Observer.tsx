import { CustomFormItem } from '@/components/CustomFormItem';
import IconTip from '@/components/IconTip';
import SelectWithTooltip from '@/components/SelectWithTooltip';
import { MINIMAL_CONFIG, SUFFIX_UNIT } from '@/constants';
import { MIRROR_SERVER } from '@/constants/doc';
import { intl } from '@/utils/intl';
import {
  Button,
  Card,
  Checkbox,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Space,
  Tooltip,
} from 'antd';
import { FormInstance } from 'antd/lib/form';
import { clone } from 'lodash';
import styles from './index.less';

interface ObserverProps {
  storageClasses: API.TooltipData[] | undefined;
  form: FormInstance<API.CreateClusterData>;
  setPvcValue: boolean;
  pvcValue: boolean;
}

const observerToolTipText = intl.formatMessage({
  id: 'OBDashboard.Cluster.New.Observer.TheImageShouldBeFully',
  defaultMessage:
    '镜像应写全 registry/image:tag，例如 oceanbase/oceanbase-cloud-native:4.2.0.0-101000032023091319',
});

export const TooltipItemContent = ({ item }) => {
  return (
    <ul style={{ margin: 0, padding: '10px' }}>
      {item.toolTipData.map((data: any) => {
        const key = Object.keys(data)[0];
        if (typeof data[key] === 'string') {
          return (
            <li style={{ listStyle: 'none' }} key={key}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                }}
              >
                <p>{key}：</p>
                <p>{data[key]}</p>
              </div>
            </li>
          );
        } else {
          const value = JSON.stringify(data[key]) || String(data[key]);
          return (
            <li style={{ listStyle: 'none' }} key={key}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                }}
              >
                <p>{key}：</p>
                <p>{value}</p>
              </div>
            </li>
          );
        }
      })}
    </ul>
  );
};

export default function Observer({
  storageClasses,
  form,
  pvcValue,
  setPvcValue,
}: ObserverProps) {
  const setMinimalConfiguration = () => {
    const originObserver = clone(form.getFieldsValue());
    form.setFieldsValue({
      ...originObserver,
      observer: {
        ...originObserver.observer,
        resource: {
          cpu: MINIMAL_CONFIG.cpu,
          memory: MINIMAL_CONFIG.memory,
        },
        storage: {
          ...originObserver.observer.storage,
          data: {
            size: MINIMAL_CONFIG.data,
          },
          log: {
            size: MINIMAL_CONFIG.log,
          },
          redoLog: {
            size: MINIMAL_CONFIG.redoLog,
          },
        },
      },
    });
  };

  return (
    <Col span={24}>
      <Card
        title="Observer"
        extra={
          <Button type="primary" onClick={setMinimalConfiguration}>
            {intl.formatMessage({
              id: 'Dashboard.Cluster.New.Observer.MinimumSpecificationConfiguration',
              defaultMessage: '最小规格配置',
            })}
          </Button>
        }
      >
        <Tooltip title={observerToolTipText}>
          <CustomFormItem
            style={{ width: '50%' }}
            message={intl.formatMessage({
              id: 'Dashboard.Cluster.New.Observer.EnterAnImage',
              defaultMessage: '请输入镜像',
            })}
            label={
              <>
                {intl.formatMessage({
                  id: 'Dashboard.Cluster.New.Observer.Image',
                  defaultMessage: '镜像',
                })}{' '}
                <a href={MIRROR_SERVER} rel="noreferrer" target="_blank">
                  {intl.formatMessage({
                    id: 'Dashboard.Cluster.New.Observer.ImageList',
                    defaultMessage: '（镜像列表）',
                  })}
                </a>
              </>
            }
            name={['observer', 'image']}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Observer.EnterAnImage',
                defaultMessage: '请输入镜像',
              })}
            />
          </CustomFormItem>
        </Tooltip>
        <Row>
          <Col>
            <p className={styles.titleText}>
              {intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Observer.Resources',
                defaultMessage: '资源',
              })}
            </p>

            <div className={styles.resourceContent}>
              <CustomFormItem
                className={styles.leftContent}
                label="cpu"
                name={['observer', 'resource', 'cpu']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.cpu}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <CustomFormItem
                label={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.Observer.Memory',
                  defaultMessage: '内存',
                })}
                name={['observer', 'resource', 'memory']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.memory}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
            </div>
          </Col>
        </Row>
        <p className={styles.titleText}>storage</p>
        <p className={styles.subTitleText}>存储卷配置</p>
        <Space style={{ marginBottom: '10px' }}>
          <IconTip
            tip={
              '勾选后，在删除 OBServer 资源之后不会级联删除 PVC；默认会进行级联删除'
            }
            content={'PVC 独立生命周期'}
          />
          <Checkbox
            value={pvcValue}
            onChange={(e) => {
              setPvcValue(e.target.checked);
            }}
          />
        </Space>
        <Row gutter={16}>
          <Col span={8}>
            <p className={styles.subTitleText}>
              {intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Observer.Data',
                defaultMessage: '数据',
              })}
            </p>
            <div className={styles.dataContent}>
              <CustomFormItem
                className={styles.leftContent}
                label="size"
                name={['observer', 'storage', 'data', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.data}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <Form.Item
                label="storageClass"
                name={['observer', 'storage', 'data', 'storageClass']}
              >
                <SelectWithTooltip
                  type="observer"
                  name={['observer', 'storage', 'data', 'storageClass']}
                  form={form}
                  selectList={storageClasses}
                  TooltipItemContent={TooltipItemContent}
                />
              </Form.Item>
            </div>
          </Col>
          <Col span={8}>
            <p className={styles.subTitleText}>
              {intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Observer.Log',
                defaultMessage: '日志',
              })}
            </p>
            <div className={styles.logContent}>
              <CustomFormItem
                className={styles.leftContent}
                label="size"
                name={['observer', 'storage', 'log', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.log}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <Form.Item
                label="storageClass"
                name={['observer', 'storage', 'log', 'storageClass']}
              >
                <SelectWithTooltip
                  type="observer"
                  name={['observer', 'storage', 'data', 'storageClass']}
                  form={form}
                  selectList={storageClasses}
                  TooltipItemContent={TooltipItemContent}
                />
              </Form.Item>
            </div>
          </Col>
          <Col span={8}>
            <p className={styles.subTitleText}>redoLog</p>
            <div className={styles.redologContent}>
              <CustomFormItem
                className={styles.leftContent}
                label="size"
                name={['observer', 'storage', 'redoLog', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.redoLog}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <Form.Item
                label="storageClass"
                validateTrigger="onBlur"
                name={['observer', 'storage', 'redoLog', 'storageClass']}
              >
                <SelectWithTooltip
                  type="observer"
                  name={['observer', 'storage', 'data', 'storageClass']}
                  form={form}
                  selectList={storageClasses}
                  TooltipItemContent={TooltipItemContent}
                />
              </Form.Item>
            </div>
          </Col>
        </Row>
      </Card>
    </Col>
  );
}
