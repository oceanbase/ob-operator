import styles from './index.less';

interface MoreModalProps {
  visible: boolean;
  ItemClick: (value: API.ModalType) => void;
  list: { value: string; label: string }[];
  innerRef: any;
  disable: boolean;
}

export default function MoreModal({
  visible,
  list,
  ItemClick,
  innerRef,
  disable,
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
              disable
                ? { color: 'rgba(0, 0, 0, 0.45)', cursor: 'not-allowed' }
                : {}
            }
            onClick={
              !disable ? () => ItemClick(item.value as API.ModalType) : () => {}
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
