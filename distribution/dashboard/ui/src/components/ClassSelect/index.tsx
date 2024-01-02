import { intl } from '@/utils/intl';
import { Select, Tooltip } from 'antd';

interface ClassSelectProps {
  selectList: {
    value: string;
    label: string;
    toolTipData: any;
  }[];
  form: any;
  name: string | number | (string | number)[];
}

export default function ClassSelect({
  selectList,
  form,
  name,
}: ClassSelectProps) {
  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) =>{
    return (option?.value ?? '').toLowerCase().includes(input.toLowerCase());
  }  

  const formatData = (list: any) => {
    list.forEach((item: any) => {
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
          title={
            <ul style={{ margin: 0, padding: '10px' }}>
              {item.toolTipData.map((data: any) => {
                let key = Object.keys(data)[0];
                if (typeof data[key] === 'string') {
                  return (
                    <li style={{ listStyle: 'none' }} key={key}>
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                        }}
                      >
                        <p>{key}：</p>
                        <p>{data[key]}</p>
                      </div>
                    </li>
                  );
                } else {
                  let value = JSON.stringify(data[key]) || String(data[key]);
                  return (
                    <li style={{ listStyle: 'none' }} key={key}>
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                        }}
                      >
                        <p>{key}：</p>
                        <p>{value}</p>
                      </div>
                    </li>
                  );
                }
              })}
            </ul>
          }
        >
          <div>{item.value}</div>
        </Tooltip>
      );
    });
    return list;
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
