import { Tooltip } from 'antd';

export default function CustomTooltip({
  text,
  width,
}: {
  text: string;
  width: number;
}) {
  return (
    <Tooltip title={text}>
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
