import MoreModal from '@/components/moreModal';
import { intl } from '@/utils/intl';
import G6,{ IG6GraphEvent } from '@antv/g6';
import { createNodeFromReact } from '@antv/g6-react-node';
import { useParams } from '@umijs/max';
import { useRequest,useUpdateEffect } from 'ahooks';
import { Spin,message } from 'antd';
import _ from 'lodash';
import { ReactElement,useEffect,useMemo,useRef,useState } from 'react';

import showDeleteConfirm from '@/components/customModal/DeleteModal';
import OperateModal from '@/components/customModal/OperateModal';
import { RESULT_STATUS } from '@/constants';
import BasicInfo from '@/pages/Cluster/Detail/Overview/BasicInfo';
import {
getClusterFromTenant,
getOriginResourceUsages,
getZonesOptions,
} from '@/pages/Tenant/helper';
import { getClusterDetailReq } from '@/services';
import {
deleteClusterReportWrap,
deleteObzoneReportWrap,
} from '@/services/reportRequest/clusterReportReq';
import { deleteObtenantPool } from '@/services/tenant';
import { ReactNode,config } from './G6register';
import type { OperateTypeLabel } from './constants';
import {
clusterOperate,
clusterOperateOfTenant,
getClusterOperates,
getZoneOperateOfCluster,
getZoneOperateOfTenant,
serverOperate,
} from './constants';
import { appenAutoShapeListener,checkIsSame,getServerNumber } from './helper';
import styles from './index.less'

interface TopoProps {
  tenantReplicas?: API.ReplicaDetailType[];
  namespace?: string;
  clusterNameOfKubectl?: string; // k8s resource name
  header?: ReactElement;
  resourcePoolDefaultValue?: any;
  refreshTenant?: () => void;
  defaultUnitCount?: number;
  status?: string;
  loading?: boolean;
}

//Cluster topology diagram component
export default function TopoComponent({
  tenantReplicas,
  header,
  namespace,
  clusterNameOfKubectl,
  resourcePoolDefaultValue,
  refreshTenant,
  defaultUnitCount,
  status,
  loading
}: TopoProps) {
  const { ns:urlNs, name:urlName } = useParams();
  const clusterOperateList = tenantReplicas
    ? clusterOperateOfTenant
    : clusterOperate;
  const modelRef = useRef<HTMLInputElement>(null);

  const [visible, setVisible] = useState<boolean>(false);
  const [operateList, setOprateList] =
    useState<OperateTypeLabel>(clusterOperateList);
  const [inNode, setInNode] = useState<boolean>(false);
  const [inModal, setInModal] = useState<boolean>(false);
  const [[ns, name]] = useState(
    namespace && clusterNameOfKubectl
      ? [namespace, clusterNameOfKubectl]
      : [urlNs, urlName],
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
  let { data: originTopoData, run: getTopoData, loading:clusterTopoLoading } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onBefore: () => {
        beforeTopoData.current = originTopoData?.topoData;
      },
    },
  );
  const clusterStatus = useRef(originTopoData?.basicInfo?.status);
  const tenantStatus = useRef(status);

  //Node more icon click event
  const handleClick = (evt: IG6GraphEvent) => {
    if (modelRef.current) {
      switch (evt.item?._cfg?.model?.type) {
        case 'cluster':
          let disabled = tenantReplicas
            ? clusterStatus.current !== 'running' ||
              tenantStatus.current !== 'running'
            : clusterStatus.current !== 'running';
          setOprateList(
            getClusterOperates(clusterOperateList, disabled),
          );
          break;
        case 'zone':
          const zone = evt.item?._cfg?.model?.label as string;
          if (tenantReplicas) {
            const { setEditZone } = resourcePoolDefaultValue;
            if (setEditZone) setEditZone(zone);
            const haveResourcePool = !!tenantReplicas?.find(
              (replica) => replica.zone === zone,
            );
            setOprateList(
              getZoneOperateOfTenant(
                haveResourcePool,
                tenantReplicas,
                tenantStatus.current,
                clusterStatus.current
              ),
            );
          } else {
            setOprateList(
              getZoneOperateOfCluster(
                originTopoData?.topoData,
                clusterStatus.current,
              ),
            );
          }
          chooseZoneName.current = zone;
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
    const res = await deleteClusterReportWrap({ ns, name });
    if (res.successful) {
      message.success(res.message);
      getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
    }
  };
  //delete zone
  const zoneDelete = async () => {
    const res = await deleteObzoneReportWrap({
      ns,
      name,
      zoneName: chooseZoneName.current,
    });
    if (res.successful) {
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.components.TopoComponent.DeletedSuccessfully',
            defaultMessage: '删除成功',
          }),
      );
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
  // delete resource pool
  const deleteResourcePool = async (zoneName: string) => {
    const res = await deleteObtenantPool({ ns, name, zoneName });
    if (res.successful) {
      if (refreshTenant) refreshTenant();
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.Detail.Overview.Replicas.DeletedSuccessfully',
            defaultMessage: '删除成功',
          }),
      );
    }
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

    if (operate === 'editResourcePools') {
      modalType.current = 'editResourcePools';
      setOperateModalVisible(true);
    }

    if (operate === 'createResourcePools') {
      modalType.current = 'createResourcePools';
      setOperateModalVisible(true);
    }

    if (operate === 'deleteResourcePool') {
      modalType.current = 'deleteResourcePool';
      showDeleteConfirm({
        onOk: () => deleteResourcePool(chooseZoneName.current),
        title: intl.formatMessage(
          {
            id: 'Dashboard.components.TopoComponent.AreYouSureYouWant',
            defaultMessage:
              '确定要删除该租户在{{chooseZoneNameCurrent}}上的资源池吗？',
          },
          { chooseZoneNameCurrent: chooseZoneName.current },
        ),
      });
    }
  };

  const mouseEnter = () => setInModal(true);
  const mouseLeave = () => setInModal(false);
  // //Re-acquire data after successful operation and maintenance operations
  // const operateSuccess = () => {
  //   getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
  // };

  //Used to re-render the view after data update
  useUpdateEffect(() => {
    let checkStatusTimer: NodeJS.Timer;
    clusterStatus.current = originTopoData?.basicInfo?.status;
    //polling
    if (!RESULT_STATUS.includes(originTopoData.topoData.status)) {
      checkStatusTimer = setInterval(() => {
        getTopoData({ ns, name, useFor: 'topo', tenantReplicas });
      }, 3000);
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

  useUpdateEffect(() => {
    tenantStatus.current = status;
  }, [status]);

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
  const isCreateResourcePool = modalType.current === 'createResourcePools';
  
  // Use different pictures for nodes in different states
  return (
    <div style={{ position: 'relative', height: '100vh' }}>
      {header
        ? header
        : originTopoData && (
            <BasicInfo
              extra={false}
              style={{ backgroundColor: '#f5f8fe', border: 'none' }}
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
          />
        ),

        [operateList, visible, status],
      )}

      <OperateModal
        type={modalType.current}
        visible={operateModalVisible}
        setVisible={setOperateModalVisible}
        successCallback={refreshTenant}
        params={{
          zoneName: chooseZoneName.current,
          defaultValue: chooseServerNum,
          defaultUnitCount: defaultUnitCount,
          ...resourcePoolDefaultValue,
          essentialParameter: isCreateResourcePool
            ? resourcePoolDefaultValue?.essentialParameter
            : getOriginResourceUsages(
                resourcePoolDefaultValue?.essentialParameter,
                resourcePoolDefaultValue?.replicaList?.find(
                  (replica) =>
                    replica.zone === resourcePoolDefaultValue.editZone,
                ),
              ),
          newResourcePool: isCreateResourcePool,
          zonesOptions: isCreateResourcePool
            ? getZonesOptions(
                getClusterFromTenant(
                  resourcePoolDefaultValue?.clusterList,
                  resourcePoolDefaultValue?.clusterResourceName,
                ),
                resourcePoolDefaultValue?.replicaList,
              )
            : undefined,
        }}
      />
      <Spin
        spinning={Boolean(clusterTopoLoading || loading)}
        size="large"
        className={styles.topoSpin}
      />
    </div>
  );
}
