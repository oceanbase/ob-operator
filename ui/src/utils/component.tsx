import { intl } from '@/utils/intl';
import { SearchOutlined } from '@ant-design/icons';
import { Button, Input, Space, type TableColumnType } from 'antd';
import { trim } from 'lodash';
import Highlighter from 'react-highlight-words';

export const getColumnSearchProps = ({
  dataIndex,
  searchInput,
  setSearchText,
  setSearchedColumn,
  searchText,
  searchedColumn,
  arraySearch,
  symbol,
}): TableColumnType<DataType> => ({
  filterDropdown: ({
    setSelectedKeys,
    selectedKeys,
    confirm,
    clearFilters,
  }) => (
    <div style={{ padding: 8 }} onKeyDown={(e) => e.stopPropagation()}>
      <Input
        ref={searchInput}
        value={selectedKeys[0]}
        onChange={(e) =>
          setSelectedKeys(e.target.value ? [e.target.value] : [])
        }
        onPressEnter={() => {
          confirm();
          setSearchText(selectedKeys[0]);
          setSearchedColumn(dataIndex);
        }}
        style={{ marginBottom: 8, display: 'block' }}
      />
      <Space>
        <Button
          type="primary"
          onClick={() => {
            confirm();
            setSearchText(selectedKeys[0]);
            setSearchedColumn(dataIndex);
          }}
          icon={<SearchOutlined />}
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
              setSearchText('');
            }
            confirm({ closeDropdown: false });
            setSearchText((selectedKeys as string[])[0]);
            setSearchedColumn(dataIndex);
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
  ),
  filterIcon: (filtered: boolean) => (
    <SearchOutlined style={{ color: filtered ? '#1677ff' : undefined }} />
  ),
  onFilter: (value, record) => {
    const safeValue = value ? value.split(symbol)[0] : '';
    const realValue = trim(safeValue.toLowerCase());

    return arraySearch
      ? record[dataIndex] &&
          record[dataIndex].some(
            (item) =>
              (item.key && item.key.toLowerCase().includes(realValue)) ||
              (item.value && item.value.toLowerCase().includes(realValue)),
          )
      : record[dataIndex] &&
          record[dataIndex]
            .toString()
            .toLowerCase()
            .includes(trim(value ? value.toLowerCase() : ''));
  },
  filterDropdownProps: {
    onOpenChange(open) {
      if (open) {
        setTimeout(() => searchInput.current?.select(), 100);
      }
    },
  },
  render: (text) =>
    searchedColumn === dataIndex ? (
      <Highlighter
        highlightStyle={{ backgroundColor: '#ffffff', padding: 0 }}
        searchWords={[searchText]}
        autoEscape
        textToHighlight={text ? text.toString() : ''}
      />
    ) : (
      text
    ),
});
