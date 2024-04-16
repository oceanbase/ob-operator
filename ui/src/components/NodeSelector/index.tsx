import { intl } from '@/utils/intl';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Col, Form, Row, Select } from 'antd';
import _ from 'lodash';

import { getNodeLabelsReq } from '@/services';
import { useEffect, useRef, useState } from 'react';

type ListType = { label: string; value: string };
interface NodeSelectorProps {
  showLabel: boolean;
  formName: string | number | (string | number)[];
  getNowNodeSelector: () => { key: string; value: string }[];
}

export default function NodeSelector({
  showLabel,
  formName,
  getNowNodeSelector,
}: NodeSelectorProps) {
  const [keyList, setKeyList] = useState<ListType[]>([]);
  const [valList, setValList] = useState<ListType[]>([]);
  const originLabels = useRef<{ key: string; value: string }[]>([]);

  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase());

  const dataFormat = (list: string[]): ListType[] => {
    const res = [];
    for (const item of list) {
      res.push({ label: item, value: item });
    }
    return res;
  };

  const checkSeletorIsExist = (nodeSelector: any): boolean => {
    for (const item of nodeSelector) {
      if (item && item.key) {
        return true;
      }
    }
    return false;
  };

  const getOriginKeys = (): string[] => {
    return _.uniq(originLabels.current.map((label: any) => label.key));
  };
  const getOriginVals = (): string[] => {
    return _.uniq(originLabels.current.map((label: any) => label.value));
  };
  // 获取value对应的keys
  const getAvailabelKeys = (data: string): string[] => {
    const res = [];
    for (const label of originLabels.current) {
      if (label.value === data) {
        res.push(label.key);
      }
    }
    return res;
  };
  //获取key对应的values
  const getAvailabelValues = (data: string): string[] => {
    const res = [];
    for (const label of originLabels.current) {
      if (label.key === data) {
        res.push(label.value);
      }
    }
    return res;
  };
  // 1、检查同组zone里面的key,可选项中的key是同组里面唯一的
  // 2、检查对应的 key或者value的值，如果有的话可选项中的value或者key是其对应的一些值
  const handleFocusKey = (selectorIdx: number, type: 'key' | 'value') => {
    //获取nodeSelector作为参数传进来
    // const topologyData = form.getFieldValue(fieldValue);
    // const { nodeSelector } = topologyData[zoneIdx]; //nodeSelector:[{key:'',value:''}]
    const nodeSelector = getNowNodeSelector();
    if (type === 'key') {
      let keys = getOriginKeys(); //key最后要格式化为 [{label:'',value:''}]
      //检查同一zone下面有哪些key已经选了
      if (checkSeletorIsExist(nodeSelector)) {
        const selectKey: string[] = nodeSelector.map((item: any) =>
          item ? item.key : '',
        );
        keys = keys.filter((key) => {
          return !selectKey.includes(key);
        });
      }
      // const { value } = nodeSelector[selectorIdx];
      const selector = nodeSelector[selectorIdx];
      let availabelKeys: string[] = [];
      if (selector && selector.value) {
        availabelKeys = getAvailabelKeys(selector.value);
        keys = keys.filter((key) => availabelKeys.includes(key));
      }
      // keys = _.difference(keys, availabelKeys);
      setKeyList(dataFormat(keys));
    } else {
      // const { key } = nodeSelector[selectorIdx];
      const selector = nodeSelector[selectorIdx];
      let values = getOriginVals();
      let availabelValues: string[] = [];
      if (selector && selector.key) {
        availabelValues = getAvailabelValues(selector.key);
        values = values.filter((val) => availabelValues.includes(val));
      }
      setValList(dataFormat(values));
    }
  };
  useEffect(() => {
    const promise = getNodeLabelsReq();
    promise.then((data) => {
      setKeyList(data.key);
      setValList(data.value);
      originLabels.current = data.originLabels;
    });
  }, []);
  return (
    <Form.Item label={showLabel && 'nodeSelector'}>
      <Form.List name={formName}>
        {(subFields, { add: subAdd, remove: subRemove }) => (
          <>
            {subFields.map(({ key, name, ...restField }, selectorIdx) => (
              <Row key={key} gutter={8}>
                <Col span={11}>
                  <Form.Item
                    {...restField}
                    name={[name, 'key']}
                    rules={[
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'OBDashboard.components.NodeSelector.EnterAKey',
                          defaultMessage: '请输入key',
                        }),
                      },
                    ]}
                  >
                    <Select
                      onFocus={() =>
                        handleFocusKey(
                          //   index,
                          selectorIdx,
                          'key',
                        )
                      }
                      showSearch
                      placeholder={intl.formatMessage({
                        id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                        defaultMessage: '请选择',
                      })}
                      optionFilterProp="label"
                      //@ts-expect-error Custom option component type is incompatible
                      filterOption={filterOption}
                      options={keyList}
                      allowClear
                    />
                  </Form.Item>
                </Col>
                :
                <Col span={11}>
                  <Form.Item
                    {...restField}
                    name={[name, 'value']}
                    rules={[
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'OBDashboard.components.NodeSelector.EnterValue',
                          defaultMessage: '请输入value',
                        }),
                      },
                    ]}
                  >
                    <Select
                      onFocus={() =>
                        handleFocusKey(
                          selectorIdx,
                          'value',
                        )
                      }
                      showSearch
                      placeholder={intl.formatMessage({
                        id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                        defaultMessage: '请选择',
                      })}
                      optionFilterProp="label"
                      //@ts-expect-error Custom option component type is incompatible
                      filterOption={filterOption}
                      options={valList}
                      allowClear
                    />
                  </Form.Item>
                </Col>
                <Col
                  style={{
                    lineHeight: '32px',
                    textAlign: 'center',
                  }}
                >
                  <MinusCircleOutlined onClick={() => subRemove(name)} />
                </Col>
              </Row>
            ))}
            <Row gutter={8}>
              <Col span={22}>
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => subAdd()}
                    block
                    icon={<PlusOutlined />}
                  >
                    {intl.formatMessage({
                      id: 'OBDashboard.components.NodeSelector.AddNodeselector',
                      defaultMessage: '添加 nodeSelector',
                    })}
                  </Button>
                </Form.Item>
              </Col>
            </Row>
          </>
        )}
      </Form.List>
    </Form.Item>
  );
}
