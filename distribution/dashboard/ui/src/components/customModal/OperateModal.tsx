import type { CommonModalType } from '.';
import AddZoneModal from './AddZoneModal';
import ScaleModal from './ScaleModal';
import UpgradeModal from './UpgradeModal';

interface OperateModalProps {
  type: API.modalType;
  zoneName?: any;
  defaultValue?:number
  disabled?:boolean
}

export default function OperateModal({
  type,
  ...props
}: OperateModalProps & CommonModalType) {
  if (type === 'addZone') {
    return <AddZoneModal {...props} />;
  }

  if (type === 'scaleServer' && props.zoneName) {
    return <ScaleModal {...props} />;
  }

  if (type === 'upgrade') {
    return <UpgradeModal {...props} />;
  }

  return;
}
