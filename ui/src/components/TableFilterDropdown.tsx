import { intl } from '@/utils/intl';
import { Button, Input, Space } from '@oceanbase/design';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';
import { useEffect, useState } from 'react';

export interface TableFilterDropdownProps extends FilterDropdownProps {
  /* 确定筛选后的回调函数 */
  onConfirm?: (value: React.Key) => void;
}

const TableFilterDropdown: React.FC<TableFilterDropdownProps> = ({
  setSelectedKeys,
  selectedKeys,
  confirm,
  clearFilters,
  visible,
  onConfirm,
}) => {
  const [value, setValue] = useState('');

  const confirmFilter = () => {
    confirm();
    if (onConfirm) {
      onConfirm(selectedKeys && selectedKeys[0]);
    }
  };

  useEffect(() => {
    if (!visible) {
      confirmFilter();
    }
  }, [visible]);

  return (
    <div
      style={{ padding: 8 }}
      onBlur={() => {
        confirmFilter();
      }}
    >
      <Input
        onChange={(e) => {
          setValue(e.target.value);
          setSelectedKeys(e.target.value ? [e.target.value] : []);
        }}
        value={value}
        onPressEnter={() => {
          confirmFilter();
        }}
        style={{ width: 188, marginBottom: 8, display: 'block' }}
      />

      <Space>
        <Button
          type="primary"
          onClick={() => {
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
        >
          {intl.formatMessage({
            id: 'src.components.333BA2EC',
            defaultMessage: '搜索',
          })}
        </Button>
        <Button
          onClick={() => {
            if (clearFilters) {
              clearFilters();
              setValue('');
            }
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
        >
          {intl.formatMessage({
            id: 'src.components.CA97A806',
            defaultMessage: '重置',
          })}
        </Button>
      </Space>
    </div>
  );
};

export default TableFilterDropdown;
