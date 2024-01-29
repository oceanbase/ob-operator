import type { CommonModalType } from '.';
import ActivateTenant from './ActivateTenantModal';
import AddZoneModal from './AddZoneModal';
import LogReplayModal from './LogReplayModal';
import ModifyPasswordModal from './ModifyPasswordModal';
import ModifyUnitModal from './ModifyUnitModal';
import ScaleModal from './ScaleModal';
import SwitchTenantModal from './SwitchTenantModal';
import UpgradeModal from './UpgradeModal';
import UpgradeTenantModal from './UpgradeTenantModal';

interface OperateModalProps {
  type: API.ModalType;
  zoneName?: any;
  defaultValue?: number;
  disabled?: boolean;
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

  if (type === 'modifyUnit') {
    return <ModifyUnitModal {...props} />;
  }

  if (type === 'changePassword') {
    return <ModifyPasswordModal {...props} />;
  }

  if (type === 'logReplay') {
    return <LogReplayModal {...props} />;
  }
  if (type === 'activateTenant') {
    return <ActivateTenant {...props} />;
  }
  if (type === 'switchTenant') {
    return <SwitchTenantModal {...props} />;
  }

  if (type === 'upgradeTenant') {
    return <UpgradeTenantModal {...props} />;
  }

  return;
}
