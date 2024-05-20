<script>

import MainHeader from '/src/components/MainHeader.vue'
import PhotoCard from '/src/components/PhotoCard.vue'

export default {
  data() {
    return {
      errormsg: '',
      username: '',
      userid: '',
      stream: null
    }
  },
  components: {
    MainHeader,
    PhotoCard
  },
  mounted() {
    this.username = localStorage.getItem('username')
    this.userid = localStorage.getItem('userid')
    if (!this.username || !this.userid) {
      this.$router.push('/login')
    }

    /* get users stream */
    this.getStream()
  },
  methods: {
    async getStream() {
      try {
        const response = await this.$axios.get(`/users/${this.userid}/stream`, {
					headers: {
						Authorization: "Bearer " + this.userid
					}
				})
        this.stream = await response.data
      } catch (e) {
        if (e.response && e.response.status === 500) {
          this.errormsg = "An internal error occurred. Please try again later."
        } else {
          this.errormsg = e.toString()
        }
      }
    }
  }
}

</script>

<template>
  <div id="homeview" style="height: 100%;">
    <MainHeader />
    <div id="content">
      <h2 id="welcome-msg">Welcome, {{ username }}! </h2>
      <hr style="border: 1px solid #4a4a4a; margin-bottom: 20px; width: 100%;"/>
      <p v-if="errormsg" id="error-msg">{{ errormsg }}</p>
      <div v-for="photo in stream" style="margin-bottom: 15px">
        <PhotoCard :photoOwner="photo.identifier.ownerId" :photoId="photo.identifier.photoId" :user="photo.user" :date="photo.dateTime" :likes="photo.likes"/>
      </div>
    </div>
  </div>
</template>

<style scoped>

#content {
  background-color: #bebebe;
  width: calc(100% - 60px);
  min-height: calc(100% - 141px);
  display: flex;
  flex-direction: column;
  padding: 30px
}

#welcome-msg {
  font-family: Arial, sans-serif;
  margin-bottom: 2px;
  margin-top: 4px;  
  margin-left: 5px
}

#error-msg {
  margin-top: 5px; 
  color: red; 
  font-weight: bold; 
  font-family: Arial, sans-serif; 
  font-size: 14px
}

</style>