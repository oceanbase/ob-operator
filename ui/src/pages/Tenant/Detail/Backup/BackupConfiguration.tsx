import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { BACKUP_RESULT_STATUS } from '@/constants';
import { usePublicKey } from '@/hook/usePublicKey';
import {
  deleteBackupReportWrap,
  editBackupReportWrap,
} from '@/services/reportRequest/backupReportReq';
import { intl } from '@/utils/intl';
import { useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  InputNumber,
  Row,
  Space,
  Typography,
  message,
} from 'antd';
import dayjs from 'dayjs';
import { useRef, useState } from 'react';
import {
  checkIsSame,
  checkScheduleDatesHaveFull,
  formatBackupForm,
  formatBackupPolicyData,
} from '../../helper';
import BakMethodsList from '../NewBackup/BakMethodsList';
import SchduleSelectFormItem from '../NewBackup/SchduleSelectFormItem';
import ScheduleTimeFormItem from '../NewBackup/ScheduleTimeFormItem';
interface BackupConfigurationProps {
  backupPolicy: API.BackupPolicy;
  setBackupPolicy: React.Dispatch<
    React.SetStateAction<API.BackupPolicy | undefined>
  >;

  backupPolicyRefresh: () => void;
}

const { Text } = Typography;

export default function BackupConfiguration({
  backupPolicy,
  setBackupPolicy,
  backupPolicyRefresh,
}: BackupConfigurationProps) {
  const [form] = Form.useForm();
  const access = useAccess();
  const scheduleValue = Form.useWatch(['scheduleDates'], form);
  const [isEdit, setIsEdit] = useState<boolean>(false);
  const curConfig = useRef({});
  const { ns, name } = useParams();
  const publicKey = usePublicKey();

  const INFO_CONFIG_ARR = [
    {
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupConfiguration.Status',
        defaultMessage: '状态',
      }),
      value: 'status',
    },
    {
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupConfiguration.BackupMediaType',
        defaultMessage: '备份介质类型',
      }),
      value: 'destType',
    },
    {
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupConfiguration.LogArchivePath',
        defaultMessage: '日志归档路径',
      }),
      value: 'archivePath',
    },
    {
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupConfiguration.DataBackupPath',
        defaultMessage: '数据备份路径',
      }),
      value: 'bakDataPath',
    },
  ];
  if (backupPolicy.ossAccessSecret) {
    INFO_CONFIG_ARR.splice(2, 0, {
      value: 'ossAccessSecret',
      label: 'OSS Access Secret',
    });
  }
  const DATE_CONFIG = {
    jobKeepDays: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.BackupConfiguration.BackupTaskRetention',
      defaultMessage: '备份任务保留',
    }),
    recoveryDays: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.BackupConfiguration.DataRecoveryWindow',
      defaultMessage: '数据恢复窗口',
    }),
    pieceIntervalDays: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.BackupConfiguration.ArchiveSliceInterval',
      defaultMessage: '归档切片间隔',
    }),
  };

  const initialValues = {
    ...backupPolicy,
    scheduleDates: {
      ...formatBackupPolicyData(backupPolicy),
    },
    scheduleTime: backupPolicy.scheduleTime
      ? dayjs(backupPolicy.scheduleTime, 'HH:mm')
      : '',
  };

  const { run: deleteBackupPolicyReq } = useRequest(deleteBackupReportWrap, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        backupPolicyRefresh();
      }
    },
  });

  const changeStatus = async () => {
    const param = {
      ns,
      name,
      status: backupPolicy.status === 'PAUSED' ? 'RUNNING' : 'PAUSED',
    };
    const { successful, data } = await editBackupReportWrap(param);
    if (successful) {
      if (data.status === backupPolicy.status) {
        backupPolicyRefresh();
      } else {
        message.success(
          intl.formatMessage({
            id: 'Dashboard.Detail.Backup.BackupConfiguration.OperationSucceeded.1',
            defaultMessage: '操作成功！',
          }),
        );
      }
    }
  };

  const changeEditBtnStatus = () => {
    if (!isEdit) {
      setIsEdit(!isEdit);
      return;
    }

    if (
      checkIsSame(
        formatBackupForm(initialValues, publicKey),
        formatBackupForm(form.getFieldsValue(), publicKey),
      )
    ) {
      message.info(
        intl.formatMessage({
          id: 'Dashboard.Detail.Backup.BackupConfiguration.NoConfigurationChangeDetected',
          defaultMessage: '未检测到配置更改',
        }),
      );
      setIsEdit(!isEdit);
      return;
    }

    form.submit();
  };

  const updateBackupPolicyConfig = async (values) => {
    if (!checkScheduleDatesHaveFull(values.scheduleDates)) {
      message.warning(
        intl.formatMessage({
          id: 'Dashboard.Detail.Backup.BackupConfiguration.ConfigureAtLeastOneFull',
          defaultMessage: '请至少配置 1 个全量备份',
        }),
      );
      return;
    }
    const { successful, data } = await editBackupReportWrap({
      ns,
      name,
      ...formatBackupForm(values, publicKey),
    });
    if (successful) {
      curConfig.current = formatBackupForm(form.getFieldsValue(), publicKey);
      // Updates are asynchronous
      if (!BACKUP_RESULT_STATUS.includes(data.status)) {
        backupPolicyRefresh();
      }
      setBackupPolicy(data);
      setIsEdit(false);
      message.success(
        intl.formatMessage({
          id: 'Dashboard.Detail.Backup.BackupConfiguration.OperationSucceeded.2',
          defaultMessage: '操作成功!',
        }),
      );
    }
  };

  return (
    <Card
      title={intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupConfiguration.BackupPolicyConfiguration',
        defaultMessage: '备份策略配置',
      })}
      style={{ width: '100%' }}
      extra={
        access.obclusterwrite ? (
          <Space>
            <Button type="primary" onClick={changeEditBtnStatus}>
              {isEdit
                ? intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.UpdateConfiguration',
                    defaultMessage: '更新配置',
                  })
                : intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.Edit',
                    defaultMessage: '编辑',
                  })}
            </Button>
            <Button
              disabled={
                backupPolicy.status !== 'RUNNING' &&
                backupPolicy.status !== 'PAUSED'
              }
              onClick={changeStatus}
            >
              {backupPolicy.status === 'PAUSING' ||
              backupPolicy.status === 'PAUSED'
                ? intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.Recovery',
                    defaultMessage: '恢复',
                  })
                : intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.Pause',
                    defaultMessage: '暂停',
                  })}
            </Button>
            <Button
              type="primary"
              danger
              onClick={() =>
                showDeleteConfirm({
                  onOk: () => deleteBackupPolicyReq({ ns, name }),
                  title: intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.AreYouSureYouWant',
                    defaultMessage: '确定要删除该备份策略吗？',
                  }),
                })
              }
            >
              {intl.formatMessage({
                id: 'Dashboard.Detail.Backup.BackupConfiguration.Delete',
                defaultMessage: '删除',
              })}
            </Button>
          </Space>
        ) : null
      }
    >
      <Form
        form={form}
        onFinish={updateBackupPolicyConfig}
        initialValues={initialValues}
      >
        <Row style={{ marginBottom: 24 }} gutter={[12, 12]}>
          {INFO_CONFIG_ARR.map((infoItem, index) => (
            <Col key={index} span={8}>
              <span style={{ marginRight: 8, color: '#8592AD', flexShrink: 0 }}>
                {infoItem.label}:
              </span>
              <Text
                style={{ width: 316 }}
                ellipsis={{ tooltip: backupPolicy[infoItem.value] }}
              >
                {backupPolicy[infoItem.value]}
              </Text>
            </Col>
          ))}
        </Row>
        <Row>
          <Col span={12}>
            <SchduleSelectFormItem
              disable={!isEdit || !access.obclusterwrite}
              form={form}
              scheduleValue={scheduleValue}
            />
          </Col>
          <Col span={12}>
            <ScheduleTimeFormItem disable={!isEdit || !access.obclusterwrite} />
          </Col>
          <Col span={12}>
            <BakMethodsList
              disable={!isEdit || !access.obclusterwrite}
              form={form}
            />
          </Col>
          <Col span={24}>
            <Descriptions>
              {backupPolicy.bakEncryptionSecret && (
                <Descriptions.Item
                  label={intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.EncryptPasswordInformation',
                    defaultMessage: '加密密码信息',
                  })}
                >
                  {backupPolicy.bakEncryptionSecret}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Col>

          {isEdit || !access.obclusterwrite ? (
            Object.keys(DATE_CONFIG).map((key, index) => (
              <Col span={8} key={index}>
                <Form.Item label={DATE_CONFIG[key]} name={key}>
                  <InputNumber
                    addonAfter={intl.formatMessage({
                      id: 'Dashboard.Detail.Backup.BackupConfiguration.Days',
                      defaultMessage: '天',
                    })}
                    min={0}
                  />
                </Form.Item>
              </Col>
            ))
          ) : (
            <Descriptions>
              {Object.keys(DATE_CONFIG).map((key, index) => (
                <Descriptions.Item label={DATE_CONFIG[key]} key={index}>
                  {backupPolicy[key]}
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.Backup.BackupConfiguration.Days',
                    defaultMessage: '天',
                  })}
                </Descriptions.Item>
              ))}
            </Descriptions>
          )}
        </Row>
      </Form>
    </Card>
  );
}
