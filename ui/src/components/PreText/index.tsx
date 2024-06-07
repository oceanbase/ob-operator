import { Tooltip } from 'antd';
import styles from './index.less';

interface PreTextProps {
  cols?: number;
  value: object | string;
}

export default function PreText({ cols, value }: PreTextProps) {
  return (
    <Tooltip
      title={
        <pre className={styles.tooltipContent}>
          {typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
        </pre>
      }
    >
      <pre
        className={cols ? styles.preText : ''}
        style={{ WebkitLineClamp: cols }}
      >
        {typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
      </pre>
    </Tooltip>
  );
}
