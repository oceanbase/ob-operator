import { Tooltip } from 'antd';

export default function CustomTooltip({
  text,
  width,
  tooltipTitle,
}: {
  text: string;
  width: number;
}) {
  return (
    <Tooltip title={tooltipTitle || text}>
      <p
        style={{
          overflow: 'hidden',
          whiteSpace: 'nowrap',
          textOverflow: 'ellipsis',
          wordBreak: 'keep-all',
          width: `${width}px`,
        }}
      >
        {text}
      </p>
    </Tooltip>
  );
}
