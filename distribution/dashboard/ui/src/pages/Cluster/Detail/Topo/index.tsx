import MoreModal from '@/components/moreModal';
import { intl } from '@/utils/intl';
import G6, { IG6GraphEvent } from '@antv/g6';
import { createNodeFromReact } from '@antv/g6-react-node';
import { useRequest, useUpdateEffect } from 'ahooks';
import { message } from 'antd';
import _ from 'lodash';
import { ReactElement, useEffect, useMemo, useRef, useState } from 'react';

import showDeleteConfirm from '@/components/customModal/DeleteModal';
import OperateModal from '@/components/customModal/OperateModal';
import { deleteObcluster, deleteObzone, getClusterDetailReq } from '@/services';
import BasicInfo from '../Overview/BasicInfo';
import { getNSName } from '../Overview/helper';
import { ReactNode, config } from './G6register';
import type { OperateType } from './constants';
import { clusterOperate, serverOperate, zoneOperate } from './constants';
import { appenAutoShapeListener, checkIsSame, getServerNumber } from './helper';

interface TopoProps {
  tenantTopoData?: API.ReplicaDetailType[];
  HeaderComp?: ReactElement;
}

export default function Topo({ tenantTopoData, HeaderComp }: TopoProps) {
  const modelRef = useRef<HTMLInputElement>(null);
  const [visible, setVisible] = useState<boolean>(false);
  const [operateList, setOprateList] = useState<OperateType>(clusterOperate);
  const [inNode, setInNode] = useState<boolean>(false);
  const [inModal, setInModal] = useState<boolean>(false);
  const [operateDisable, setOperateDisable] = useState<boolean>(false);
  //当前集群ns [ns,name]
  const [[ns, name]] = useState(getNSName());
  //控制运维弹窗显隐
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  //当前运维弹窗类型
  const modalType = useRef<API.ModalType>('addZone');
  //当前点击的节点id
  const currentId = useRef<string>('');
  const graph = useRef<any>(null);
  const beforeTopoData = useRef<any>(null);
  //点击 more icon 时选中的 zone name
  const chooseZoneName = useRef<string>('');
  // 选中zone下的server数量
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  //topoData集群状态如果是operating 需要轮询
  let { data: originTopoData, run: getTopoData } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onBefore: () => {
        beforeTopoData.current = originTopoData?.topoData;
      },
    },
  );
  //节点 more icon 点击事件
  const handleClick = (evt: IG6GraphEvent) => {
    if (modelRef.current) {
      switch (evt.item?._cfg?.model?.type) {
        case 'cluster':
          setOprateList(clusterOperate);
          break;
        case 'zone':
          setOprateList(zoneOperate);
          chooseZoneName.current = evt.item?._cfg?.model?.label as string;
          break;
        case 'server':
          setOprateList(serverOperate);
          break;
      }
      currentId.current = evt.item!._cfg!.id as string;
      setVisible(true);
      modelRef.current.style.left = `${evt.canvasX + 5}px`;
      modelRef.current.style.top = `${evt.clientY - 40}px`;
    }
  };
  //删除集群
  const clusterDelete = async () => {
    const res = await deleteObcluster({ ns, name });
    if (res.successful) {
      message.success(res.message);
      getTopoData({ ns, name, useFor: 'topo', tenantTopoData });
    }
  };
  //删除Zone
  const zoneDelete = async () => {
    const res = await deleteObzone({
      ns,
      name,
      zoneName: chooseZoneName.current,
    });
    if (res.successful) {
      message.success(res.message);
      getTopoData({ ns, name, useFor: 'topo', tenantTopoData });
    }
  };
  //初始化g6
  const init = () => {
    const container = document.getElementById('topoContainer');
    const width = container?.scrollWidth || 1280;
    const height = container?.scrollHeight || 500;

    graph.current = new G6.TreeGraph(config(width, height));
    G6.registerNode('cluster', createNodeFromReact(ReactNode(handleClick)));
    G6.registerNode('zone', createNodeFromReact(ReactNode(handleClick)));
    G6.registerNode('server', createNodeFromReact(ReactNode()));
    G6.registerEdge('flow-line', {
      draw(cfg, group) {
        const startPoint = cfg.startPoint!;
        const endPoint = cfg.endPoint!;
        const { style } = cfg;
        const shape = group.addShape('path', {
          attrs: {
            stroke: style!.stroke,
            //箭头endArrow
            // endArrow: style.endArrow,
            // path 线条路径，可以是 String 形式，也可以是线段的数组。格式参考：SVG path。
            path: [
              // M移动画笔 不画线 L划线
              ['M', startPoint.x, startPoint.y], //M: Move to 需要移动到对应点的x、y坐标
              ['L', startPoint.x, (startPoint.y + endPoint.y) / 2], // L:line to
              ['L', endPoint.x, (startPoint.y + endPoint.y) / 2],
              ['L', endPoint.x, endPoint.y],
            ],
          },
        });

        return shape;
      },
    });

    /**
     * 有一些事件的注册会有异常，比如mouseleave，可以在此处手动监听注册
     */
    graph.current.on('node:mouseleave', () => {
      setInNode(false);
    });
    graph.current.on('node:mouseenter', () => {
      setInNode(true);
    });
    graph.current.data(_.cloneDeep(originTopoData.topoData)); //防止graph修改原有数据
    graph.current.render();
    appenAutoShapeListener(graph.current);
  };

  /**
   * 调起运维操作弹窗
   */
  const ItemClickOperate = (operate: string) => {
    if (operate === 'addZone') {
      modalType.current = 'addZone';
      setOperateModalVisible(true);
    }
    if (operate === 'upgradeCluster') {
      modalType.current = 'upgrade';
      setOperateModalVisible(true);
    }

    if (operate === 'deleteCluster') {
      showDeleteConfirm({
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Topo.AreYouSureYouWant',
          defaultMessage: '你确定要删除集群吗？',
        }),
        onOk: clusterDelete,
      });
    }
    if (operate === 'scaleServer') {
      modalType.current = 'scaleServer';
      setChooseServerNum(
        getServerNumber(originTopoData.topoData, chooseZoneName.current),
      );
      setOperateModalVisible(true);
    }
    if (operate === 'deleteZone') {
      showDeleteConfirm({
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Topo.AreYouSureYouWant.1',
          defaultMessage: '你确定要删除该Zone吗？',
        }),
        onOk: zoneDelete,
      });
    }
  };

  const mouseEnter = () => setInModal(true);
  const mouseLeave = () => setInModal(false);
  //运维操作成功后重新获取数据
  const operateSuccess = () => {
    getTopoData({ ns, name, useFor: 'topo', tenantTopoData });
  };

  //用于数据更新后重新渲染视图
  useUpdateEffect(() => {
    let checkStatusTimer: NodeJS.Timer;
    //轮询
    if (originTopoData.topoData.status === 'operating') {
      if (!operateDisable) setOperateDisable(true);
      checkStatusTimer = setInterval(() => {
        getTopoData({ ns, name, useFor: 'topo', tenantTopoData });
      }, 3000);
    } else {
      if (operateDisable) setOperateDisable(false);
    }
    if (graph.current) {
      if (!checkIsSame(beforeTopoData.current, originTopoData.topoData)) {
        let _topoData = _.cloneDeep(originTopoData.topoData);
        beforeTopoData.current = _topoData;
        graph.current.changeData(_topoData);
      }
    } else {
      init();
    }
    return () => {
      if (checkStatusTimer) clearInterval(checkStatusTimer);
    };
  }, [originTopoData]);

  /**
   * 弹窗打开时会触发mouseleave，因此不能在mouseleave的回调函数中关闭modal
   *
   * 解决方法：
   * 维护两个变量inModa、inNode
   * 当两个变量都为false时 即鼠标即不在modal内也不在node内 弹窗隐藏
   *
   * 缺点：变量变化太频繁 性能不友好
   */
  useEffect(() => {
    if (!inModal && !inNode) {
      setVisible(false);
      currentId.current = '';
    }
  }, [inModal, inNode]);

  useEffect(() => {
    if (modelRef.current) {
      modelRef.current.addEventListener('mouseenter', mouseEnter);
      modelRef.current.addEventListener('mouseleave', mouseLeave);
    }
    getTopoData({ ns, name, useFor: 'topo', tenantTopoData });

    return () => {
      modelRef.current?.removeEventListener('mouseenter', mouseEnter);
      modelRef.current?.removeEventListener('mouseleave', mouseLeave);
    };
  }, []);

  // 针对不同状态的node 使用不同的图片
  return (
    <div style={{ position: 'relative', height: '100vh' }}>
      {HeaderComp
        ? HeaderComp
        : originTopoData && (
            <BasicInfo
              style={{ backgroundColor: '#f5f8fe' }}
              {...(originTopoData.basicInfo as API.ClusterInfo)}
            />
          )}
      <div style={{ height: '100%' }} id="topoContainer"></div>
      {useMemo(
        () => (
          <MoreModal
            id={currentId.current}
            innerRef={modelRef}
            visible={visible}
            list={operateList}
            ItemClick={ItemClickOperate}
            disable={operateDisable}
          />
        ),

        [operateDisable, visible],
      )}

      <OperateModal
        type={modalType.current}
        visible={operateModalVisible}
        setVisible={setOperateModalVisible}
        successCallback={operateSuccess}
        zoneName={chooseZoneName.current}
        defaultValue={chooseServerNum}
      />
    </div>
  );
}
