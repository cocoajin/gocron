import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import store from './store/index'
const app = createApp(App)
app.use(ElementPlus, { locale: zhCn, size: 'default' })
app.use(router)
app.use(store)
app.directive('focus', { mounted: el => el.focus() })
app.config.globalProperties.$appConfirm = function (callback) {
 this.$confirm('确定执行此操作?', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
 .then(() => callback()).catch(() => {})
}
app.config.globalProperties.$formatTime = function (time) {
  const fillZero = function (num) {
    return num >= 10 ? num : '0' + num
  }
  const date = new Date(time)

  const result = date.getFullYear() + '-' +
  (fillZero(date.getMonth() + 1)) + '-' +
  fillZero(date.getDate()) + ' ' +
  fillZero(date.getHours()) + ':' +
  fillZero(date.getMinutes()) + ':' +
  fillZero(date.getSeconds())

  if (result.indexOf('20') !== 0) {
    return ''
  }

  return result
}

app.mount('#app')
