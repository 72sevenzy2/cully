import DefaultTheme from 'vitepress/theme'
import InstallCommand from './InstallCommand.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('InstallCommand', InstallCommand)
  }
}
