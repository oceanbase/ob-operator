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
    <Card title="详细配置">
      <Row>
        <p className={styles.titleText}>资源设置</p>
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
            name="serviceType"
            label="服务类型"
          >
            <Select placeholder="请选择" options={SERVICE_TYPE} />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem name="replicas" label="副本数">
            <InputNumber placeholder="请输入" min={1} />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem name={['resource', 'cpu']} label="CPU 核数">
            <InputNumber placeholder="请输入" min={1} />
          </CustomFormItem>
        </Col>
        <Col span={8}>
          <CustomFormItem name={['resource', 'memory']} label="内存大小">
            <InputNumber
              placeholder="请输入"
              min={1}
              addonAfter={SUFFIX_UNIT}
            />
          </CustomFormItem>
        </Col>
      </Row>
      <Row>
        <p className={styles.titleText}>参数设置</p>
        <Col span={24}>
          <CustomFormItem name={'parameters'}> 
            <InputLabelComp />
          </CustomFormItem>
        </Col>
      </Row>
    </Card>
  );
}
