// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "WhatsApp Cloud API",
  tagline: "Rest api for whatsapp communication.",
  url: "https://api.codechat.dev/",
  baseUrl: "/",
  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/logo.png",
  organizationName: "CodeChat",
  projectName: "CodeChat",

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          docLayoutComponent: "@theme/DocPage",
          docItemComponent: "@theme/ApiItem",
        },
        blog: {
          showReadingTime: true,
          blogTitle: "Change Log",
          blogDescription: "CodeChat Change Log",
          blogSidebarTitle: "Records",
          path: "./change-log",
          feedOptions: {
            title: "Change Log",
            type: "all"
          },
          routeBasePath: "/change-log"
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        gtag: {
          trackingID: "GTM-THVM29S",
          anonymizeIP: false,
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      docs: {
        sidebar: {
          hideable: true,
          autoCollapseCategories: false
        },
      },
      navbar: {
        title: "CodeChat",
        logo: {
          alt: "CodeChat",
          src: "img/logo.png",
          href: "/"
        },
        items: [
          // {
          //   type: 'docSidebar',
          //   sidebarId: 'tutorialSidebar',
          //   position: 'left',
          //   label: 'Eventos',
          // },
          {
            label: "API",
            position: "left",
            to: "/",
          },
          {
            label: "Webhook",
            position: "left",
            to: "/webhook/v1.0.0",
          },
          // {
          //   type: 'dropdown',
          //   label: "Api Versions",
          //   position: "left",
          //   items: [
          //     {
          //       label: 'v3.0.0',
          //       to: "/api/v3.0.0",
          //     },
          //   ]
          // },
          {
            label: 'ChangeLog',
            to: '/change-log',
            position: 'right',
          },
          {
            href: "https://github.com/code-chat-br",
            position: "right",
            className: "header-github-link",
            "aria-label": "GitHub repository",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "Docs",
            items: [
              {
                label: "CodeChat Docs",
                to: "/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "WhatsApp Group",
                href: "https://chat.whatsapp.com/HyO8X8K0bAo0bfaeW8bhY5",
              },
              {
                label: "Telegram Group",
                href: "https://t.me/codechatBR"
              }
            ],
          },
          {
            title: 'Company',
            items: [
              {
                label: 'Terms of Service',
                to: '/term-of-use'
              }
            ]
          }
        ],
        copyright: `Copyright © ${new Date().getFullYear()} CodeChat, Inc.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ["ruby", "csharp", "php"],
      },
      languageTabs: [
        {
          highlight: "bash",
          language: "curl",
          logoClass: "bash",
        },
        {
          highlight: "python",
          language: "python",
          logoClass: "python",
        },
        {
          highlight: "go",
          language: "go",
          logoClass: "go",
        },
        {
          highlight: "javascript",
          language: "nodejs",
          logoClass: "nodejs",
        },
        {
          highlight: "ruby",
          language: "ruby",
          logoClass: "ruby",
        },
        {
          highlight: "csharp",
          language: "csharp",
          logoClass: "csharp",
        },
        {
          highlight: "php",
          language: "php",
          logoClass: "php",
        }
      ],
      algolia: {
        appId: "4UO6XBSILO",
        apiKey: "5caaa17365282ffa9092f32b8d022c7d",
        indexName: "codechat",
        inputSelector: "Search docs",
        contextualSearch: true
      }
    }),

  plugins: [
    [
      "docusaurus-plugin-openapi-docs",
      {
        id: "openapi",
        docsPluginId: "classic",
        config: {
          openapi_versioned: {
            specPath: "openapi/versions/v3.0.0.yaml",
            outputDir: "docs/openapi/v3.0.0",
            sidebarOptions: {
              groupPathsBy: "tag",
              categoryLinkSource: "tag",
            },
            version: "3.0.0",
            label: "v3.0.0",
            baseUrl: "/api/v3.0.0",
            versions: {},
          },
          'webhook-v1.0.0': {
            specPath: "openapi/webhook/v1.0.0.yaml",
            outputDir: "docs/webhook/v1.0.0",
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: "tag",
              categoryLinkSource: "tag",
            },
            version: "1.0.0",
            label: "v1.0.0",
            baseUrl: "/webhook/v1.0.0",
            versions: {},
          },
        },
      },
    ],
    [
      "@docusaurus/plugin-pwa",
      {
        debug: true,
        offlineModeActivationStrategies: [
          "appInstalled",
          "standalone",
          "queryString",
        ],
      },
    ],
  ],
  themes: ["docusaurus-theme-openapi-docs"],
  stylesheets: [
    {
      href: "https://use.fontawesome.com/releases/v5.11.0/css/all.css",
      type: "text/css",
    },
  ],
};

module.exports = config;
