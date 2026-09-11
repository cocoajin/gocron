<template>
  <div v-cloak>
    <el-menu
      :default-active="currentRoute"
      mode="horizontal"
      :ellipsis="false"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
      router>
      <el-menu-item index="/task">任务管理</el-menu-item>
      <el-menu-item index="/host">任务节点</el-menu-item>
      <el-menu-item v-if="$store.getters.user.isAdmin" index="/user">用户管理</el-menu-item>
      <el-menu-item v-if="$store.getters.user.isAdmin" index="/system">系统管理</el-menu-item>
      <div style="flex: 1"></div>
      <el-sub-menu v-if="$store.getters.user.token" index="userStatus">
        <template #title>{{$store.getters.user.username}}</template>
        <el-menu-item index="/user/edit-my-password">修改密码</el-menu-item>
        <el-menu-item @click="logout" index="/user/login">退出</el-menu-item>
      </el-sub-menu>

    </el-menu>
  </div>
</template>

<script>

export default {
  name: 'app-nav-menu',
  data () {
    return {}
  },
  computed: {
    currentRoute () {
      if (this.$route.path === '/') {
        return '/task'
      }
      const segments = this.$route.path.split('/')
      return `/${segments[1]}`
    }
  },
  methods: {
    logout () {
      this.$store.commit('logout')
      this.$router.push('/')
    }
  }
}
</script>
