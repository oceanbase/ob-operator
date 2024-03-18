import { intl } from '@/utils/intl';
import { Select, Tooltip } from 'antd';

interface SelectWithTooltipProps {
  selectList: API.TooltipData[];
  form: any;
  name: string | number | (string | number)[];
  TooltipItemContent: (item: API.TooltipData) => JSX.Element;
}

export default function SelectWithTooltip({
  selectList,
  form,
  name,
  TooltipItemContent
}: SelectWithTooltipProps) {
  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) =>{
    return (option?.value ?? '').toLowerCase().includes(input.toLowerCase());
  }  

  const formatData = (selectList: API.TooltipData[]) => {
    selectList.forEach((item: API.TooltipData) => {
      item.label = (
        <Tooltip
          color="#fff"
          overlayInnerStyle={{ color: 'rgba(0,0,0,.85)' }}
          overlayStyle={{
            borderRadius: '2px',
            maxWidth: '400px',
            boxShadow:
              '0 3px 6px -4px rgba(0,0,0,.12), 0 6px 16px 0 rgba(0,0,0,.08), 0 9px 28px 8px rgba(0,0,0,.05)',
          }}
          placement="bottomLeft"
          title={<TooltipItemContent item={item}/>}
        >
          <div>{item.value}</div>
        </Tooltip>
      );
    });
    return selectList;
  };

  const selectChange = (val: string) => {
    form.setFieldValue(name, val);
    form.validateFields([name]);
  };

  return (
    <Select
      showSearch
      placeholder={intl.formatMessage({
        id: 'OBDashboard.components.ClassSelect.PleaseSelect',
        defaultMessage: '请选择',
      })}
      optionFilterProp="value"
      filterOption={filterOption}
      options={formatData(selectList)}
      onChange={selectChange}
      value={form.getFieldValue(name)}
      // dropdownRender={DropDownComponent}
    />
  );
}
