import { intl } from '@/utils/intl';
import {
  Button,
  Card,
  Col,
  Input,
  InputNumber,
  Row,
  Tooltip,
} from 'antd';
import { FormInstance } from 'antd/lib/form';
import { CustomFormItem } from '@/components/CustomFormItem';
import SelectWithTooltip from '@/components/SelectWithTooltip';
import { MINIMAL_CONFIG, SUFFIX_UNIT } from '@/constants';
import { MIRROR_SERVER } from '@/constants/doc';
import { clone } from 'lodash';
import styles from './index.less';

interface ObserverProps {
  storageClasses: API.TooltipData[] | undefined;
  form: FormInstance<API.CreateClusterData>;
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

export default function Observer({ storageClasses, form }: ObserverProps) {
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
              <CustomFormItem
                label="storageClass"
                name={['observer', 'storage', 'data', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    name={['observer', 'storage', 'data', 'storageClass']}
                    form={form}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}

                {/* <CustomSelect /> */}
              </CustomFormItem>
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
              <CustomFormItem
                label="storageClass"
                name={['observer', 'storage', 'log', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    form={form}
                    name={['observer', 'storage', 'log', 'storageClass']}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}
              </CustomFormItem>
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
              <CustomFormItem
                label="storageClass"
                validateTrigger="onBlur"
                name={['observer', 'storage', 'redoLog', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    form={form}
                    name={['observer', 'storage', 'redoLog', 'storageClass']}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}
              </CustomFormItem>
            </div>
          </Col>
        </Row>
      </Card>
    </Col>
  );
}
