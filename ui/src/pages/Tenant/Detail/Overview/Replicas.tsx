import { Col, Descriptions, Card } from 'antd';

export default function Replicas({replicaList}:{replicaList: API.ReplicaDetailType[]}) {
  return (
    <Col span={24}>
      {replicaList.map((replica, index) => (
        <Card title={<h2>replicas</h2>} collapsible key={index}>
          <Descriptions column={5} title={`replica ${index + 1}`}>
            {Object.keys(replica).map((key, idx) => (
              <Descriptions.Item label={key} key={idx}>{replica[key]}</Descriptions.Item>
            ))}
          </Descriptions>
        </Card>
      ))}
    </Col>
  );
}
