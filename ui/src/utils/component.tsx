import TableFilterDropdown from '@/components/TableFilterDropdown';
import { SearchOutlined } from '@ant-design/icons';
import { token } from '@oceanbase/design';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';

export const getColumnSearchProps = ({
  frontEndSearch,
  dataIndex,
  onConfirm,
}: {
  frontEndSearch: boolean; // 前端分页时，dataIndex 必传，该参数对后端分页无效
  dataIndex?: string;
  onConfirm?: (value?: React.Key) => void;
}) => ({
  filterDropdown: (props: FilterDropdownProps) => (
    <TableFilterDropdown {...props} onConfirm={onConfirm} />
  ),

  filterIcon: (filtered: boolean) => (
    <SearchOutlined
      style={{ color: filtered ? token.colorPrimary : undefined }}
    />
  ),

  // 前端搜索，需要定义 onFilter 函数
  ...(frontEndSearch
    ? {
        onFilter: (value, record) =>
          record[dataIndex] &&
          record[dataIndex]
            .toString()
            .toLowerCase()
            .includes(value && value.toLowerCase()),
      }
    : {}),
});
