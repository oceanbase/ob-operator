import { InputNumber, Select } from 'antd';
import { useState } from 'react';

interface InputTimeCompProps {
  onChange?: (value: number | null) => void;
  value?: number;
}

type UnitType = 'second' | 'minute' | 'hour';

const SELECT_OPTIONS = [
  {
    label: '秒',
    value: 'second',
  },
  {
    label: '分钟',
    value: 'minute',
  },
  {
    label: '小时',
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

  return (
    <InputNumber
      min={1}
      onChange={durationChange}
      value={getValue(value)}
      addonAfter={
        <Select
          defaultValue={'minute'}
          onChange={(val: UnitType) => setUnit(val)}
          options={SELECT_OPTIONS}
        />
      }
    />
  );
}
