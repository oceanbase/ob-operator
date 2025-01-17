import { intl } from '@/utils/intl';
import { InputNumber, Select } from 'antd';
import { useEffect, useState } from 'react';

interface InputTimeCompProps {
  onChange?: (value: number | null) => void;
  value?: number;
}

type UnitType = 'second' | 'minute' | 'hour';

const SELECT_OPTIONS = [
  {
    label: intl.formatMessage({
      id: 'src.components.InputTimeComp.5C00F6E6',
      defaultMessage: '秒',
    }),
    value: 'second',
  },
  {
    label: intl.formatMessage({
      id: 'src.components.InputTimeComp.7D0C8989',
      defaultMessage: '分钟',
    }),
    value: 'minute',
  },
  {
    label: intl.formatMessage({
      id: 'src.components.InputTimeComp.C34DD7B4',
      defaultMessage: '小时',
    }),
    value: 'hour',
  },
];

export default function InputTimeComp({ onChange, value }: InputTimeCompProps) {
  const [unit, setUnit] = useState<UnitType>('minute');

  const durationChange = (value: number | null) => {
    if (!value || unit === 'second') onChange?.(value);
    if (unit === 'minute') {
      onChange?.(value! * 60);
    }
    if (unit === 'hour') {
      onChange?.(value! * 3600);
    }
  };
  const getValue = (value: number | undefined) => {
    if (!value) return value;
    if (unit === 'minute') {
      return Math.floor(value / 60);
    }
    if (unit === 'hour') {
      return Math.floor(value / 3600);
    }
    return value;
  };

  useEffect(() => {
    if (value && value < 60 && unit === 'minute') {
      setUnit('second');
    }
  }, [value]);

  return (
    <InputNumber
      min={1}
      onChange={durationChange}
      value={getValue(value)}
      placeholder={intl.formatMessage({
        id: 'src.components.InputTimeComp.9628BB2C',
        defaultMessage: '请输入',
      })}
      addonAfter={
        <Select
          defaultValue={'minute'}
          value={unit}
          onChange={(val: UnitType) => setUnit(val)}
          options={SELECT_OPTIONS}
        />
      }
    />
  );
}
