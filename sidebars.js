/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

/**
 * Melhore o codigo abaixo para que ele gere a sidebar automaticamente.
 * Qual a sugestão?
 * 
 */

const sidebars = {
  // tutorialSidebar: [
  //   {
  //     type: "category",
  //     label: "Comunicação e Mensagens",
  //     link: {
  //       type: "generated-index"
  //     },
  //     items: [{ type: 'autogenerated', dirName: 'message_communication' }]
  //   }
  // ],

  codechat: [
    {
      type: "category",
      label: "v3.0.0",
      link: {
        type: "generated-index",
        title: "WhatsApp Cloud API",
        description:
          "Use o CodeChat para enviar e receber mensagens personalizadas em grande escala e criar automação com agentes virtuais, aprimorando a experiência do cliente e melhorando a produtividade dos funcionários.",
        slug: "/",
      },
      items: require("./docs/openapi/v3.0.0/sidebar.js"),
    }
  ],

  webhook: [
    {
      type: "category",
      label: "Webhook",
      link: {
        type: "generated-index",
        slug: "/webhook/v1.0.0",
      },
      items: require("./docs/webhook/v1.0.0/sidebar.js"),
    },
  ],
};

module.exports = sidebars;
