import styles from './index.less';

interface MoreModalProps {
  visible: boolean;
  //   setVisible: (prop: boolean) => void;
  ItemClick: (value: string, id: string) => void;
  list: { value: string; label: string }[];
  innerRef: any;
  id: string;
  disable: boolean;
}

export default function MoreModal({
  visible,
  list,
  ItemClick,
  innerRef,
  id,
  disable,
}: MoreModalProps) {
  return (
    <ul
      ref={innerRef}
      className={visible ? `${styles.moreContainer}` : `${styles.hidden}`}
    >
      {list.map((item, index) => (
        <li style={disable ? {color:'rgba(0, 0, 0, 0.45)',cursor:'not-allowed'} : {}} onClick={!disable ? () => ItemClick(item.value, id) : ()=>{}} key={index}>
          {item.label}
        </li>
      ))}
    </ul>
  );
}
