import { Button, Modal } from 'antd';

import { intl } from '@/utils/intl';
import { Terminal } from '@xterm/xterm';
import React from 'react';

export interface ITerminal {
  terminalId: string;
  onClose?: () => void;
  onConnected?: () => void;
}

function devLog(...args: any[]) {
  if (process.env.NODE_ENV === 'development') {
    console.log(...args);
  }
}

export const OBTerminal: React.FC<ITerminal> = (props) => {
  const { terminalId } = props;
  const ref = React.useRef<HTMLDivElement>(null);
  const [ws, setWs] = React.useState<WebSocket | null>(null);

  React.useEffect(() => {
    if (ref.current) {
      const term = new Terminal({
        cols: 160,
        rows: 60,
      });
      term.open(ref.current);

      if (term.options.fontSize) {
        const containerWidth = ref.current.clientWidth;

        const cols = Math.floor(containerWidth / 9.2);
        const rows = Math.floor(cols / 4);
        term.resize(cols, rows);

        const protocol = location.protocol === 'https:' ? 'wss' : 'ws';

        const ws = new WebSocket(
          `${protocol}://${location.host}/api/v1/terminal/${terminalId}?cols=${cols}&rows=${rows}`,
        );
        term.write('Hello from \x1B[1;3;31mOceanBase\x1B[0m\r\n');

        ws.onopen = function () {
          devLog('Websocket connection open ...');
          // ws.send(JSON.stringify({ type: 'ping' }))
          term.onData(function (data) {
            ws.send(data);
          });
          props.onConnected?.();
          setWs(ws);

          window.addEventListener('beforeunload', () => {
            if (ws) {
              ws.close();
              props.onClose?.();
              setWs(null);
            }
          });
        };

        ws.onmessage = function (event) {
          term.write(event.data);
        };

        ws.onclose = function () {
          devLog('Connection closed.');
          term.write('\r\nConnection closed.\r\n');
        };

        ws.onerror = function (evt) {
          console.error('WebSocket error observed:', evt);
        };
      }
    }

    return () => {
      window.removeEventListener('beforeunload', () => {});
    };
  }, []);

  return (
    <>
      {ws && (
        <div style={{ marginBottom: 12 }}>
          <Button
            danger
            type="primary"
            onClick={() => {
              Modal.confirm({
                title: intl.formatMessage({
                  id: 'Dashboard.components.Terminal.Disconnect',
                  defaultMessage: '断开连接',
                }),
                content: intl.formatMessage({
                  id: 'Dashboard.components.Terminal.Disconnect1',
                  defaultMessage: '确定要断开连接吗？',
                }),
                okType: 'danger',
                onOk: () => {
                  ws.close();
                  props.onClose?.();
                  setWs(null);
                },
              });
            }}
          >
            {intl.formatMessage({
              id: 'Dashboard.components.Terminal.Disconnect',
              defaultMessage: '断开连接',
            })}
          </Button>
        </div>
      )}
      <div ref={ref} />
    </>
  );
};
