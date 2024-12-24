import {
  ExclamationCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import type { TooltipProps } from 'antd';
import { Tooltip } from 'antd';
import classnames from 'classnames';
import React from 'react';
import './index.less';

interface IProps {
  content?: string | React.ReactNode;
  tip?: string | React.ReactNode;
  size?: 'small' | 'middle' | 'large';
  className?: string;
  style?: React.CSSProperties;
  contentProps?: React.HTMLAttributes<HTMLSpanElement>;
  tipProps?: TooltipProps;
  icon?: React.ReactNode;
  iconProps?: {
    style?: React.CSSProperties;
    className?: string;
  };
}

const InternalIconTip = (props: IProps) => {
  const {
    content,
    tip,
    className,
    style,
    size,
    contentProps,
    tipProps,
    iconProps,
    icon,
  } = props;

  return (
    <span
      className={classnames({
        'ob-cloud-icon-tip': true,
        small: size === 'small',
        large: size === 'large',
        className: !!className,
        [className]: !!className,
      })}
      style={style}
    >
      <span {...contentProps}> {content}</span>
      <Tooltip title={tip} {...tipProps}>
        {icon || <QuestionCircleOutlined {...iconProps} />}
      </Tooltip>
    </span>
  );
};

const BaseTip = (props: Omit<IProps, 'tip' | 'tipProps'>) => {
  const { content, size, className, style, contentProps, iconProps, icon } =
    props;

  return (
    <span
      className={classnames({
        'ob-cloud-base-tip': true,
        small: size === 'small',
        large: size === 'large',
        className: !!className,
        [className]: !!className,
      })}
      style={style}
    >
      {icon || <ExclamationCircleOutlined {...iconProps} />}
      <span {...contentProps}> {content}</span>
    </span>
  );
};

type InternalIconTipType = typeof InternalIconTip;

export interface IconTipInstance extends InternalIconTipType {
  BaseTip: typeof BaseTip;
}

const IconTip = InternalIconTip as IconTipInstance;

IconTip.BaseTip = BaseTip;

export default IconTip;
