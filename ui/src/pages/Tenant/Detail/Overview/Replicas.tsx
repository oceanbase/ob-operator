import CollapsibleCard from '@/components/CollapsibleCard';
import { Col,Descriptions } from 'antd';

export default function Replicas({
  replicaList,
}: {
  replicaList: API.ReplicaDetailType[];
}) {
  return (
    <Col span={24}>
      <CollapsibleCard
        title={<h2 style={{ marginBottom: 0 }}>replicas</h2>}
        collapsible={true}
        defaultExpand={true}
      >
        {replicaList.map((replica, index) => (
          <Descriptions key={index} column={5} title={`replica ${index + 1}`}>
            {Object.keys(replica).map((key, idx) => (
              <Descriptions.Item label={key} key={idx}>
                {replica[key]}
              </Descriptions.Item>
            ))}
          </Descriptions>
        ))}
      </CollapsibleCard>
    </Col>
  );
}
