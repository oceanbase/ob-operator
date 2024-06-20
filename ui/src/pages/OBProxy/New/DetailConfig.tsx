import { CustomFormItem } from '@/components/CustomFormItem';
import InputLabelComp from '@/components/InputLabelComp';
import { SERVICE_TYPE, SUFFIX_UNIT } from '@/constants';
import { MIRROR_OBPROXY } from '@/constants/doc';
import { intl } from '@/utils/intl';
import { Card, Col, Input, InputNumber, Row, Select } from 'antd';
import styles from './index.less';

const commonStyle = { width: 280 };

export default function DetailConfig() {
  return (
    <Card
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.New.7CAF48E9',
        defaultMessage: '详细配置',
      })}
    >
      <Row>
        <p className={styles.titleText}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.New.0C4EFBB0',
            defaultMessage: '资源设置',
          })}
        </p>
        <Col span={24}>
          <CustomFormItem
            style={{ width: '50%' }}
            label={
              <>
                {intl.formatMessage({
                  id: 'Dashboard.Cluster.New.Observer.Image',
                  defaultMessage: '镜像',
                })}{' '}
                <a href={MIRROR_OBPROXY} rel="noreferrer" target="_blank">
                  {intl.formatMessage({
                    id: 'Dashboard.Cluster.New.Observer.ImageList',
                    defaultMessage: '（镜像列表）',
                  })}
                </a>
              </>
            }
            name="image"
            message={intl.formatMessage({
              id: 'Dashboard.Cluster.New.Observer.EnterAnImage',
              defaultMessage: '请输入镜像',
            })}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Observer.EnterAnImage',
                defaultMessage: '请输入镜像',
              })}
            />
          </CustomFormItem>
        </Col>
        <Col span={24}>
          <CustomFormItem
            style={commonStyle}
            message={intl.formatMessage({
              id: 'src.pages.OBProxy.New.CCD9785D',
              defaultMessage: '请选择服务类型',
            })}
            name="serviceType"
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.88D0BC94',
              defaultMessage: '服务类型',
            })}
          >
            <Select
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.2F497A97',
                defaultMessage: '请选择',
              })}
              options={SERVICE_TYPE}
            />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem
            name="replicas"
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.A3E900B4',
              defaultMessage: '副本数',
            })}
          >
            <InputNumber
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.D4645164',
                defaultMessage: '请输入',
              })}
              min={1}
            />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem
            name={['resource', 'cpu']}
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.6A1E93D2',
              defaultMessage: 'CPU 核数',
            })}
          >
            <InputNumber
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.AEDDBA86',
                defaultMessage: '请输入',
              })}
              min={1}
            />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem
            name={['resource', 'memory']}
            label={intl.formatMessage({
              id: 'src.pages.OBProxy.New.CE387455',
              defaultMessage: '内存大小',
            })}
          >
            <InputNumber
              placeholder={intl.formatMessage({
                id: 'src.pages.OBProxy.New.7C04AD55',
                defaultMessage: '请输入',
              })}
              min={1}
              addonAfter={SUFFIX_UNIT}
            />
          </CustomFormItem>
        </Col>
      </Row>
      <Row>
        <p className={styles.titleText}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.New.134CD1CE',
            defaultMessage: '参数设置',
          })}
        </p>
        <Col span={24}>
          <CustomFormItem
            rules={[
              {
                required: false,
              },
            ]}
            name={'parameters'}
          >
            <InputLabelComp />
          </CustomFormItem>
        </Col>
      </Row>
    </Card>
  );
}
