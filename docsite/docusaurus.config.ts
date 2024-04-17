import { themes as prismThemes } from 'prism-react-renderer'
import type { Config } from '@docusaurus/types'
import type * as Preset from '@docusaurus/preset-classic'

const config: Config = {
  title: 'ob-operator',
  tagline: 'ob-operator is a Kubernetes operator for OceanBase',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://oceanbase.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/ob-operator',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'oceanbase', // Usually your GitHub org/user name.
  projectName: 'ob-operator', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh-Hans'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/oceanbase/ob-operator/tree/master/docsite',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  trailingSlash: false,
  themeConfig: {
    algolia: {
      appId: '6JQM9QDU5V',
      apiKey: '75f5591a502e47777a08a02b96bc09a1',
      indexName: 'oceanbaseio',
      contextualSearch: false,
      searchPagePath: false,
      // @ts-ignore
      maxResultsPerGroup: 20,
    },
    image: 'img/logo.png',
    navbar: {
      title: 'ob-operator',
      logo: {
        alt: 'OceanBase Logo',
        src: 'img/logo.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'manualSidebar',
          label: 'Manual',
          position: 'left',
        },
        {
          type: 'docSidebar',
          sidebarId: 'developerSidebar',
          label: 'Developer',
          position: 'left',
        },
        {
          label: 'Change log',
          to: 'changelog',
        },
        {
          type: 'localeDropdown',
          position: 'right',
        },
        {
          href: 'https://github.com/oceanbase/ob-operator',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Manual',
              to: '/docs/manual/what-is-ob-operator',
            },
            {
              label: 'Architecture',
              to: '/docs/developer/arch',
            },
            {
              label: 'Development',
              to: '/docs/developer/develop-locally',
            }
          ],
        },
        {
          title: 'Repos',
          items: [
            {
              label: 'ob-operator',
              href: 'https://github.com/oceanbase/ob-operator',
            },
            {
              label: 'OceanBase CE',
              href: 'https://github.com/oceanbase/oceanbase',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Forum (in Chinese)',
              href: 'https://ask.oceanbase.com/',
            },
            {
              label: 'Slack',
              href: 'https://oceanbase.slack.com/',
            },
            {
              label: 'Stack Overflow',
              href: 'https://stackoverflow.com/questions/tagged/oceanbase',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} OceanBase, Inc. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
}

export default config
