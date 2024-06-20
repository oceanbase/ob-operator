/*
 * Copyright 2023 OceanBase
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const path = require('path');
const pkg = require('./package.json');

const baseDir = path.join(__dirname, '..');

const localePath = path.join(baseDir, 'ui/src/i18n');

const outputPath = path.join(localePath, './strings');
const exclude = 'src/main';


function matchText(text, path) {
  const isConsoleLog = /^console\.log\(/gi.test(path?.parentPath?.toString());
  let isFormattedMessage = false;
  try {
    isFormattedMessage = /^\<FormattedMessage/g.test(
      path.parentPath.parentPath.parentPath.parentPath.parentPath.toString(),
    );
  } catch (e) {}
  return (
    /[\u{4E00}-\u{9FFF}]+(?![\u3000-\u303F\uFF01-\uFF5E])/gimu.test(text) &&
    !isConsoleLog
  );
}

const config = {
  name: pkg.name,
  entry: 'src',
  output: outputPath,
  sep: '.',
  exclude: (path) => {
    return (
      path.includes('src/.umi') ||
      path.includes('src/locales') ||
      (!!exclude && path.includes(exclude))
    );
  },
  sourceLang: 'zh-CN',
  targetLang: 'en-US',
  clearLangs: ['zh-CN', 'en-US'],
  matchFunc: matchText,
  injectContent: {
    import: "import { intl } from '@/utils/intl';\n",
    method: `intl.formatMessage({id: '$key$' })`,
    withDefaultMessage: true,
    defaultMessageKey: 'defaultMessage',
  },
};

module.exports = config;
