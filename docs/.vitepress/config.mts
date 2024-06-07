import {defineConfig} from 'vitepress'
import taskLists from 'markdown-it-task-lists'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "LagrangeGo 文档",
  description: "LagrangeGo 开发文档",
  base: "/LagrangeGo/",
  cleanUrls: true,
  lang: "zh-cn",
  markdown: {
    config: (md) => {
      md.use(taskLists)
    }
  },
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      {text: '简介', link: '/guide'},
      {text: '示例', link: '/examples'}
    ],

    sidebar: [
      {
        text: '简介',
        collapsed: false,
        items: [
          {text: 'LagrangeGo', link: '/guide'}
        ]
      },
      {
        text: '示例',
        collapsed: false,
        items: [
          {text: '创建bot实例', link: '/examples/createClient'},
          {text: '登录', link: '/examples/login'},
        ]
      },
    ],

    socialLinks: [
      {icon: 'github', link: 'https://github.com/LagrangeDev/LagrangeGo'}
    ],
    outline: {
      label: '目录'
    },
    docFooter: {
      prev: '上一页',
      next: '下一页'
    },
    editLink: {
      pattern: 'https://github.com/LagrangeDev/LagrangeGo/edit/master/docs/:path'
    },
    lightModeSwitchTitle: "切换到浅色模式",
    darkModeSwitchTitle: "切换到深色模式",
    lastUpdated: {
      text: '上次编辑'
    }
  }
})
