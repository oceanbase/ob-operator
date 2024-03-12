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
import BasicInfo from '@/pages/Cluster/Detail/Overview/BasicInfo';
import { deleteObcluster, deleteObzone, getClusterDetailReq } from '@/services';
import { getNSName } from '../../pages/Cluster/Detail/Overview/helper';
import { ReactNode, config } from './G6register';
import type { OperateTypeLabel } from './constants';
import { RESULT_STATUS } from '@/constants';
import {
  clusterOperate,
  clusterOperateOfTenant,
  serverOperate,
  zoneOperate,
  zoneOperateOfTenant,
} from './constants';
import { appenAutoShapeListener, checkIsSame, getServerNumber } from './helper';

interface TopoProps {
  tenantReplicas?: API.ReplicaDetailType[];
  namespace?: string;
  clusterNameOfKubectl?: string; // k8s resource name
  header?: ReactElement;
}

//Cluster topology diagram component
export default function TopoComponent({
  tenantReplicas,
  header,
  namespace,
  clusterNameOfKubectl,
}: TopoProps) {
  const clusterOperateList = tenantReplicas
    ? clusterOperateOfTenant
    : clusterOperate;
  const zoneOperateList = tenantReplicas ? zoneOperateOfTenant : zoneOperate;
  const modelRef = useRef<HTMLInputElement>(null);
  const [visible, setVisible] = useState<boolean>(false);
  const [operateList, setOprateList] =
    useState<OperateTypeLabel>(clusterOperateList);
  const [inNode, setInNode] = useState<boolean>(false);
  const [inModal, setInModal] = useState<boolean>(false);
  const [operateDisable, setOperateDisable] = useState<boolean>(false);
  const [[ns, name]] = useState(
    namespace && clusterNameOfKubectl
      ? [namespace, clusterNameOfKubectl]
      : getNSName(),
  );
  //Control the visibility of operation and maintenance modal
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  //Current operation and maintenance modal type
  const modalType = useRef<API.ModalType>('addZone');
  //The currently clicked node id
  const currentId = useRef<string>('');
  const graph = useRef<any>(null);
  const beforeTopoData = useRef<any>(null);
  //The zone name selected when clicking the more icon
  const chooseZoneName = useRef<string>('');
  //Number of servers in the selected zone
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  //If the topoData cluster status is operating, it needs to be polled.
  let { data: originTopoData, run: getTopoData } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onBefore: () => {
        beforeTopoData.current = originTopoData?.topoData;
      },
    },
  );
  //Node more icon click event
  const handleClick = (evt: IG6GraphEvent) => {
    if (modelRef.current) {
      switch (evt.item?._cfg?.model?.type) {
        case 'cluster':
          setOprateList(clusterOperateList);
          break;
        case 'zone':
          setOprateList(zoneOperateList);
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
  //delete cluster
  const clusterDelete = async () => {
    const res = await deleteObcluster({ ns, name });
    if (res.successful) {
      message.success(res.message);
      getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
    }
  };
  //delete zone
  const zoneDelete = async () => {
    const res = await deleteObzone({
      ns,
      name,
      zoneName: chooseZoneName.current,
    });
    if (res.successful) {
      message.success(res.message);
      getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
    }
  };
  //Initialize g6
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
            path: [
              ['M', startPoint.x, startPoint.y], //M: Move to
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
     * There are exceptions in the registration of some events,
     * such as mouseleave. You can manually monitor the registration here.
     */
    graph.current.on('node:mouseleave', () => {
      setInNode(false);
    });
    graph.current.on('node:mouseenter', () => {
      setInNode(true);
    });
    graph.current.data(_.cloneDeep(originTopoData.topoData)); //Prevent graph from modifying original data
    graph.current.render();
    appenAutoShapeListener(graph.current);
  };

  /**
   * Call up the operation and maintenance operation modal
   */
  const ItemClickOperate = (operate: API.ModalType) => {
    if (operate === 'addZone') {
      modalType.current = 'addZone';
      setOperateModalVisible(true);
    }
    if (operate === 'upgradeCluster') {
      modalType.current = 'upgradeCluster';
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
    if (operate === 'changeUnitCount') {
      modalType.current = 'changeUnitCount';
      setOperateModalVisible(true);
    }

    if (operate === 'modifyUnitSpecification') {
      modalType.current = 'modifyUnitSpecification';
      setOperateModalVisible(true);
    }
  };

  const mouseEnter = () => setInModal(true);
  const mouseLeave = () => setInModal(false);
  //Re-acquire data after successful operation and maintenance operations
  const operateSuccess = () => {
    getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
  };

  //Used to re-render the view after data update
  useUpdateEffect(() => {
    let checkStatusTimer: NodeJS.Timer;
    //polling
    if (!RESULT_STATUS.includes(originTopoData.topoData.status)) {
      if (!operateDisable) setOperateDisable(true);
      checkStatusTimer = setInterval(() => {
        getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
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
   * Mouseleave will be triggered when the modal is opened,
   * so the modal cannot be closed in the callback function of mouseleave.
   *
   * Solution:
   * Maintain two variables inModa and inNode
   * When both variables are false, that is,
   * the mouse is neither in the modal nor the node,
   * and the modal is hidden.
   *
   * Disadvantages: Variables change too frequently, performance is not friendly
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

    getTopoData({ ns, name, useFor: 'topo', tenantReplicas });

    return () => {
      modelRef.current?.removeEventListener('mouseenter', mouseEnter);
      modelRef.current?.removeEventListener('mouseleave', mouseLeave);
    };
  }, []);

  // Use different pictures for nodes in different states
  return (
    <div style={{ position: 'relative', height: '100vh' }}>
      {header
        ? header
        : originTopoData && (
            <BasicInfo
              style={{ backgroundColor: '#f5f8fe', border:'none' }}
              {...(originTopoData.basicInfo as API.ClusterInfo)}
            />
          )}
      <div style={{ height: '100%' }} id="topoContainer"></div>
      {useMemo(
        () => (
          <MoreModal
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
