import { intl } from '@/utils/intl';
import { PlusOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { Divider, Form, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';

import AddNSModal from '@/components/customModal/AddNSModal';
import { resourceNameRule } from '@/constants/rules';
import { getNameSpaces } from '@/services';
import { useState } from 'react';
export default function SelectNSFromItem({
  form,
}: {
  form: FormInstance<API.CreateClusterData>;
}) {
  // control the modal for adding a new namespace
  const [visible, setVisible] = useState(false);
  const { data, run: getNS } = useRequest(getNameSpaces);

  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase());
  const DropDownComponent = (menu: any) => {
    return (
      <div>
        {menu}
        <Divider style={{ margin: '10px 0' }} />
        <div
          onClick={() => setVisible(true)}
          style={{ padding: '8px', cursor: 'pointer' }}
        >
          <PlusOutlined />
          <span style={{ marginLeft: '6px' }}>
            {intl.formatMessage({
              id: 'OBDashboard.Cluster.New.BasicInfo.AddNamespace',
              defaultMessage: '新增命名空间',
            })}
          </span>
        </div>
      </div>
    );
  };
  const addNSCallback = (newNS: string) => {
    form.setFieldValue('namespace', newNS);
    form.validateFields(['namespace']);
    getNS();
  };

  return (
    <>
      <Form.Item
        label={intl.formatMessage({
          id: 'OBDashboard.Cluster.New.BasicInfo.Namespace',
          defaultMessage: '命名空间',
        })}
        name="namespace"
        validateTrigger="onBlur"
        validateFirst
        rules={[
          {
            required: true,
            message: intl.formatMessage({
              id: 'OBDashboard.Cluster.New.BasicInfo.EnterANamespace',
              defaultMessage: '请输入命名空间',
            }),
            validateTrigger: 'onChange',
          },
          resourceNameRule,
        ]}
      >
        <Select
          showSearch
          placeholder={intl.formatMessage({
            id: 'OBDashboard.Cluster.New.BasicInfo.PleaseSelect',
            defaultMessage: '请选择',
          })}
          optionFilterProp="label"
          filterOption={filterOption}
          dropdownRender={DropDownComponent}
          options={data}
        />
      </Form.Item>
      <AddNSModal
        visible={visible}
        setVisible={setVisible}
        successCallback={addNSCallback}
      />
    </>
  );
}
