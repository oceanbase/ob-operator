import TableFilterDropdown from '@/components/TableFilterDropdown';
import { SearchOutlined } from '@ant-design/icons';
import { token } from '@oceanbase/design';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';

export const getColumnSearchProps = ({
  frontEndSearch,
  dataIndex,
  onConfirm,
  arraySearch,
  symbol,
}: {
  frontEndSearch: boolean; // 前端分页时，dataIndex 必传，该参数对后端分页无效
  arraySearch: boolean; // 当前定位搜索的表头值是都为数组类型
  symbol?: string; // 拼接符号，取值为符号前，主要为支持node
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
        onFilter: (value, record) => {
          const realValue = (value && value.split(symbol)[0]).toLowerCase();

          return arraySearch
            ? record[dataIndex].some(
                (item) =>
                  item.key.toLowerCase().includes(realValue) ||
                  item.value.toLowerCase().includes(realValue),
              )
            : record[dataIndex] &&
                record[dataIndex]
                  .toString()
                  .toLowerCase()
                  .includes(value && value.toLowerCase());
        },
      }
    : {}),
});
