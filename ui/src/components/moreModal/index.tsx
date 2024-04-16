import styles from './index.less';
import { Topo } from '@/type/topo';

interface MoreModalProps {
  visible: boolean;
  ItemClick: (value: API.ModalType) => void;
  list: Topo.OperateTypeLabel;
  innerRef: any;
}

export default function MoreModal({
  visible,
  list,
  ItemClick,
  innerRef,
}: MoreModalProps) {
  return (
    <div className={styles.moreModalContainer}>
      <ul
        ref={innerRef}
        className={visible ? `${styles.moreContainer}` : `${styles.hidden}`}
      >
        {list.map((item, index) => (
          <li
            style={
              item?.disabled
                ? { color: 'rgba(0, 0, 0, 0.45)', cursor: 'not-allowed' }
                : {}
            }
            onClick={
              !item?.disabled
                ? () => ItemClick(item.value as API.ModalType)
                : () => {}
            }
            key={index}
          >
            {item.label}
          </li>
        ))}
      </ul>
    </div>
  );
}
