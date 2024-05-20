<script>

import MainHeader from '/src/components/MainHeader.vue'
import UploadPhoto from '/src/components/UploadPhoto.vue'
import ChangeUsername from '/src/components/ChangeUsername.vue'

export default {
  data() {
    return {
      username: '',
      userid: ''
    }
  },
  components: {
    MainHeader,
    UploadPhoto,
    ChangeUsername
  },
  mounted() {
    this.username = localStorage.getItem('username')
    this.userid = localStorage.getItem('userid')
    if (!this.username || !this.userid) {
      this.$router.push('/login')
    }
  }
}

</script>

<template>
  <div id="upload-view" style="height: 100%;">
    <MainHeader />
    <div id="content">
      <UploadPhoto />
      <ChangeUsername />
    </div>
  </div>
</template>

<style scoped>

#content {
  background-color: #bebebe;
  width: calc(100% - 60px);
  min-height: calc(100% - 141px);
  padding: 30px
}

#content > * {
  margin-bottom: 30px;
}

</style>