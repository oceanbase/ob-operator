import { POINT_NUMBER, REFRESH_FREQUENCY } from '@/constants';
import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { useUpdateEffect } from 'ahooks';
import { Col, DatePicker, Row, Select, Switch } from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import moment from 'moment';
import { useEffect, useState } from 'react';
import { caculateStep } from './helper';
import type {
  FilterDataType,
  Label,
  LabelType,
  OptionType,
  QueryRangeType,
} from './index';
import styles from './index.less';

interface DataFilterProps {
  isRefresh: boolean;
  realTime: string;
  filterData: FilterDataType;
  filterLabel: LabelType[];
  queryRange: QueryRangeType;
  setQueryRange: React.Dispatch<React.SetStateAction<QueryRangeType>>;
  setIsRefresh: React.Dispatch<React.SetStateAction<boolean>>;
  setFilterLable: React.Dispatch<React.SetStateAction<LabelType[]>>;
  setFilterData: React.Dispatch<React.SetStateAction<FilterDataType>>;
}
const { RangePicker } = DatePicker;
const DateSelectOption: OptionType[] = [
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.CustomTime',
      defaultMessage: '自定义时间',
    }),
    value: 'custom',
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyMinutes',
      defaultMessage: '近30分钟',
    }),
    value: 1800000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHour',
      defaultMessage: '近1小时',
    }),
    value: 3600000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours',
      defaultMessage: '近3小时',
    }),
    value: 10800000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.1',
      defaultMessage: '近6小时',
    }),
    value: 21600000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.2',
      defaultMessage: '近12小时',
    }),
    value: 43200000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.3',
      defaultMessage: '近24小时',
    }),
    value: 86400000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.LastDays',
      defaultMessage: '近7天',
    }),
    value: 604800000,
  },
];

type RangeValue = [Dayjs | null, Dayjs | null] | null;

//选择时间下拉框改变右边时间选择框
//所以右边时间选择框需要value属性 受控
export default function DataFilter({
  isRefresh,
  realTime,
  filterData,
  filterLabel, //发送请求的label
  queryRange, //defaultVAlue
  setIsRefresh,
  setFilterLable,
  setQueryRange,
}: DataFilterProps) {
  const [zoneOption, setZoneOption] = useState<OptionType[]>([]);
  const [serverOption, setServerOption] = useState<OptionType[]>([]);
  const [selectZone, setSelectZone] = useState<string>();
  const [selectServer, setSelectServer] = useState<string>();
  const [dateValue, setDateValue] = useState<RangeValue>([
    dayjs(queryRange.startTimestamp * 1000),
    dayjs(queryRange.endTimestamp * 1000),
  ]);
  const [selectRange, setSelectRange] = useState<string | number>('custom');
  const AutoRefresh = () => {
    const textStyle = {
      marginRight: 12,
      color: '#8592ad',
    };
    return (
      <div>
        {isRefresh && (
          <>
            <span style={{ ...textStyle }}>
              {intl.formatMessage(
                {
                  id: 'OBDashboard.Detail.Monitor.DataFilter.UpdateTimeRealtime',
                  defaultMessage: '更新时间：{realTime}',
                },
                { realTime: realTime },
              )}
            </span>
            <span style={{ ...textStyle }}>
              {intl.formatMessage({
                id: 'OBDashboard.Detail.Monitor.DataFilter.UpdateFrequency',
                defaultMessage: '更新频率：',
              })}
              {REFRESH_FREQUENCY}
              {intl.formatMessage({
                id: 'OBDashboard.Detail.Monitor.DataFilter.Seconds',
                defaultMessage: '秒',
              })}
            </span>
          </>
        )}
        {intl.formatMessage({
          id: 'OBDashboard.Detail.Monitor.DataFilter.AutoRefresh',
          defaultMessage: '自动刷新：',
        })}

        <Switch
          checked={isRefresh}
          onChange={(value) => {
            setIsRefresh(value);
          }}
        />
      </div>
    );
  };

  const clearLabel = (current: LabelType[], key: Label): LabelType[] => {
    let newLable = [...current];
    const idx = newLable.findIndex((item: LabelType) => item.key === key);
    if (idx !== -1) {
      newLable.splice(idx, 1);
    }
    return newLable;
  };

  //替换 或者 添加
  const updateLable = (
    current: LabelType[],
    key: Label,
    value: string,
  ): LabelType[] => {
    let newLable = [...current];
    const idx = newLable.findIndex((item: LabelType) => item.key === key);
    if (idx !== -1) {
      newLable[idx] = { key, value };
    } else {
      newLable.push({
        key,
        value,
      });
    }
    return newLable;
  };

  const handleLabel = (val: string | undefined): LabelType[] => {
    let isClear: boolean = !Boolean(val),
      currentLable = [...filterLabel];
    if (isClear) {
      //清空obzone&svr_ip
      currentLable = clearLabel(clearLabel(filterLabel, 'obzone'), 'svr_ip');
    } else {
      //更新zone后清空server
      currentLable = clearLabel(
        updateLable(filterLabel, 'obzone', val!),
        'svr_ip',
      );
    }
    return currentLable;
  };

  const zoneSelectChange = (val: string | undefined) => {
    setSelectZone(val);
    setSelectServer(undefined);
    setFilterLable(handleLabel(val));
    //清空
    if (typeof val === 'undefined') {
      setServerOption(filterData.serverList);
      return;
    }
    const filterServers = filterData.serverList.filter((server: OptionType) => {
      return server.zone === val;
    });
    setServerOption(filterServers);
  };

  const serverSelectChange = (val: string | undefined) => {
    const isClear: boolean = !Boolean(val);
    let lable: LabelType[] = [...filterLabel];
    if (isClear) {
      lable = clearLabel(lable, 'svr_ip');
    } else {
      lable = updateLable(lable, 'svr_ip', val!);
    }
    setFilterLable(lable);
    setSelectServer(val);
  };

  const range = (start: number, end: number) => {
    const result = [];
    for (let i = start; i < end; i++) {
      result.push(i);
    }
    return result;
  };

  const disabledDateTime: RangePickerProps['disabledTime'] = (_) => {
    const isToday = _?.date() === moment().date();
    if (!isToday)
      return {
        disabledHours: () => [],
        disabledMinutes: () => [],
        disabledSeconds: () => [],
      };
    return {
      disabledHours: () => range(0, 24).splice(moment().hour() + 1, 24),
      disabledMinutes: (hour) => {
        if (hour === moment().hour()) {
          return range(0, 60).splice(moment().minute() + 1, 60);
        }
        return [];
      },
      disabledSeconds: (hour, minute) => {
        if (hour === moment().hour() && minute === moment().minute()) {
          return range(0, 60).splice(moment().second(), 60);
        }
        return [];
      },
    };
  };

  const disabledDate: RangePickerProps['disabledDate'] = (current) => {
    return current && current > dayjs().endOf('day');
  };

  const rangePickerChange = (date: RangeValue) => {
    if (selectRange !== 'custom') setSelectRange('custom');
    setDateValue(date);
  };

  const selectRangeChange = (value: string | number) => {
    setSelectRange(value);
  };

  useUpdateEffect(() => {
    if (filterData.zoneList.length) {
      setZoneOption(filterData.zoneList);
    }
    if (filterData.serverList.length) {
      setServerOption(filterData.serverList);
    }
  }, [filterData]);

  useUpdateEffect(() => {
    if (selectRange !== 'custom') {
      let nowTimestamp = new Date().valueOf();
      let startTimestamp = nowTimestamp - (selectRange as number);
      setDateValue([dayjs(startTimestamp), dayjs(nowTimestamp)]);
    }
  }, [selectRange]);

  useEffect(() => {
    if (dateValue?.length) {
      const [startDate, endDate] = dateValue;

      if (startDate && endDate) {
        let startTimestamp = Math.ceil(startDate.valueOf() / 1000);
        let endTimestamp = Math.ceil(endDate.valueOf() / 1000);
        setQueryRange({
          startTimestamp,
          endTimestamp,
          step: caculateStep(startTimestamp, endTimestamp, POINT_NUMBER),
        });
      }
    }
  }, [dateValue]);

  return (
    <ProCard
      style={{ marginTop: 12 }}
      headerBordered
      title={intl.formatMessage({
        id: 'OBDashboard.Detail.Monitor.DataFilter.DataFiltering',
        defaultMessage: '数据筛选',
      })}
      extra={<AutoRefresh />}
    >
      <Row gutter={12} style={{ alignItems: 'center' }}>
        <Col span={5}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <span style={{ marginRight: 8 }}>Zone:</span>
            <Select
              value={selectZone}
              onChange={zoneSelectChange}
              allowClear
              style={{ width: '100%' }}
              showSearch
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Detail.Monitor.DataFilter.All',
                defaultMessage: '全部',
              })}
              options={zoneOption}
            />
          </div>
        </Col>
        <Col span={5}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {' '}
            <span style={{ marginRight: 8 }}>OBServer:</span>
            <Select
              value={selectServer}
              onChange={serverSelectChange}
              allowClear
              style={{ width: '100%' }}
              showSearch
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Detail.Monitor.DataFilter.All',
                defaultMessage: '全部',
              })}
              options={serverOption}
            />{' '}
          </div>
        </Col>
        <Col span={14}>
          <div
            className={styles.selectRangeTimeContainer}
            style={{ display: 'flex', alignItems: 'center' }}
          >
            <span style={{ marginRight: 8 }}>
              {intl.formatMessage({
                id: 'OBDashboard.Detail.Monitor.DataFilter.SelectTime',
                defaultMessage: '选择时间:',
              })}
            </span>
            <Select
              style={{ width: 120 }}
              value={selectRange}
              onChange={selectRangeChange}
              options={DateSelectOption}
            />

            <RangePicker
              value={dateValue}
              onChange={rangePickerChange}
              disabledDate={disabledDate}
              disabledTime={disabledDateTime}
              showTime={{ format: 'HH:mm:ss' }}
            />
          </div>
        </Col>
      </Row>
    </ProCard>
  );
}
