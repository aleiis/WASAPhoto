<script>

import MainHeader from '/src/components/MainHeader.vue'
import PhotoCard from '/src/components/PhotoCard.vue'

export default {
  props: ['profileUsername'],
  data() {
    return {
      dataLoaded: false,
      errormsg: '',
      userid: '',
      username: '',
      profileId: null,
      profile: {
        username: '',
        photos: [
          {
            ownerId: 0,
            photoId: 0
          }
        ],
        uploads: 0,
        followers: 0,
        following: 0
      },
      following: false,
      banned: false
    }
  },
  components: {
    MainHeader,
    PhotoCard
  },
  watch: {
    '$route.params.profileUsername': {
      immediate: true,
      handler() {
        this.fetchProfileData()
      }
    }
  },
  mounted() {
    this.username = localStorage.getItem('username')
    this.userid = localStorage.getItem('userid')
    if (!this.username || !this.userid) {
      this.$router.push('/login')
    }

    this.fetchProfileData()
  },
  methods: {
    async fetchProfileData() {
      this.dataLoaded = false
      await this.getProfile()
      await this.getFollow()
      await this.getBan()
      this.dataLoaded = true
    },

    /* profile */
    async getProfile() {
      try {
        const response = await this.$axios.get(`/users/?username=${this.profileUsername}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.profileId = response.data.userId
      } catch (e) {
        if (e.response && e.response.status == 404) {
          this.$router.push('/404')
        } else if (e.response && e.response.status === 500) {
          this.errormsg = "An internal error occurred. Please try again later."
        } else {
          this.errormsg = e.toString()
        }
      }

      try {
        const response = await this.$axios.get(`/users/${this.profileId}/profile`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.profile = response.data
      } catch (e) {
        if (e.response && (e.response.status == 404 || e.response.status == 403)) {
          this.$router.push('/404')
        } else if (e.response && e.response.status === 500) {
          this.errormsg = "An internal error occurred. Please try again later."
        } else {
          this.errormsg = e.toString()
        }
      }
    },

    /* follow */
    async getFollow() {
      try {
        const response = await this.$axios.get(`/users/${this.userid}/follows/${this.profileId}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        response.status === 200 ? this.following = true : this.following = false
      } catch (e) {
        if (e.response && e.response.status === 404) {
          this.following = false
        } else {
          console.error(e)
        }
      }
    },

    toggleFollow() {
      if (this.following) {
        this.unfollowUser()
      } else {
        this.followUser()
      }
    },

    async followUser() {
      try {
        await this.$axios.post(`/users/${this.userid}/follows/`, this.profileId, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.profile.followers++
        this.following = true
      } catch (e) {
        console.error(e)
      }      
    },

    async unfollowUser() {
      try {
        await this.$axios.delete(`/users/${this.userid}/follows/${this.profileId}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.profile.followers--
        this.following = false
      } catch (e) {
        console.error(e)
      }
    },

    /* ban */
    async getBan() {
      try {
        const response = await this.$axios.get(`/users/${this.userid}/bans/${this.profileId}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        response.status === 200 ? this.banned = true : this.banned = false
      } catch (e) {
        if (e.response && e.response.status === 404) {
          this.banned = false
        } else {
          console.error(e)
        }
      }
    },

    toggleBan() {
      if (this.banned) {
        this.unbanUser()
      } else {
        this.banUser()
      }
    },

    async banUser() {
      try {
        await this.$axios.post(`/users/${this.userid}/bans/`, this.profileId, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.banned = true
      } catch (e) {
        console.error(e)
      }      
    },

    async unbanUser() {
      try {
        await this.$axios.delete(`/users/${this.userid}/bans/${this.profileId}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`
          }
        })
        this.banned = false
      } catch (e) {
        console.error(e)
      }
    }
  }
}

</script>

<template>
  <div id="profile-view" style="height: 100%;">
    <MainHeader />
    <div id="content">
      <div style="display: flex; flex-direction: row; align-items: center; justify-content: space-between;">
        <div id="profile-msg">
          <h2 id="welcome-msg">{{ profileUsername }}'s profile</h2>
        </div>
        <div id="profile-interations">
          <button v-if="userid != profileId" id="follow-button" @click="toggleFollow">{{ following ? "Unfollow" : "Follow" }}</button>
          <button v-if="userid != profileId" id="ban-button" @click="toggleBan">{{ banned ? "Unban" : "Ban" }}</button>
        </div>
      </div>
      <p>{{ profile.uploads }} photos | {{ profile.followers }} followers | {{ profile.following }} following</p>
      <hr style="border: 1px solid #4a4a4a; margin-bottom: 20px; width: 100%;"/>
      <p v-if="errormsg" id="error-msg">{{ errormsg }}</p>
      <div v-if="dataLoaded" v-for="photo in profile.photos" style="margin-bottom: 15px">
        <PhotoCard @photoDeleted="this.fetchProfileData()" :photoOwner="this.profileId" :photoId="photo.photoId" :user="this.profileUsername" :date="photo.dateTime" :likes="photo.likes"/>
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

#follow-button {
  background-color: #4a4a4a;
  color: white;
  border: none;
  border-radius: 5px;
  padding: 5px 10px;
  margin-right: 15px;
  cursor: pointer
}

#ban-button {
  background-color: #4a4a4a;
  color: rgb(255, 0, 0);
  border: none;
  border-radius: 5px;
  padding: 5px 10px;
  cursor: pointer
}

</style>