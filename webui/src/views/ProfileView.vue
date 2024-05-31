<script>

import MainHeader from '/src/components/MainHeader.vue'
import PhotoCard from '/src/components/PhotoCard.vue'

export default {
	props: ['profileUsername'],
	data() {
		return {
			dataLoaded: false,
			errormsg: '',
			userId: '',
			username: '',
			profileId: null,
			profile: {
				"owner": {
					"user_id": 0,
					"username": "Maria"
				},
				"photos": [
					{
						"owner": {
							"user_id": 0,
							"username": "Maria"
						},
						"photo_id": 0,
						"date": "2019-08-24T14:15:22Z",
						"total_likes": 0,
						"total_comments": 0
					}
				],
				"uploads": 0,
				"followers": 0,
				"following": 0
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
			handler() {
				this.fetchProfileData()
			}
		}
	},
	mounted() {
		this.username = localStorage.getItem('username')
		this.userId = parseInt(localStorage.getItem('userid'), 10)
		if (!this.username || !this.userId) {
			this.$router.push('/login')
		}

		this.fetchProfileData()
	},
	methods: {
		async fetchProfileData() {
			this.dataLoaded = false
			await this.getProfile()
			if (this.profileId !== this.userId) {
				await this.getFollow()
				await this.getBan()
			}
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
				this.profileId = response.data.user_id
			} catch (e) {
				if (e.response && e.response.status === 404) {
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
				if (e.response && (e.response.status === 404 || e.response.status === 403)) {
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
				const response = await this.$axios.get(`/users/${this.userId}/follows/${this.profileId}`, {
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
				await this.$axios.post(`/users/${this.userId}/follows/`, {
					follower: parseInt(this.userId, 10),
					followed: parseInt(this.profileId, 10)
				}, {
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
				await this.$axios.delete(`/users/${this.userId}/follows/${this.profileId}`, {
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
				const response = await this.$axios.get(`/users/${this.userId}/bans/${this.profileId}`, {
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
				await this.$axios.post(`/users/${this.userId}/bans/`, {
					ban_issuer: parseInt(this.userId, 10),
					banned_user: parseInt(this.profileId, 10)
				}, {
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
				await this.$axios.delete(`/users/${this.userId}/bans/${this.profileId}`, {
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
		<MainHeader/>
		<div id="content">
			<div style="display: flex; flex-direction: row; align-items: center; justify-content: space-between;">
				<div id="profile-msg">
					<h2 id="welcome-msg">{{ profileUsername }}'s profile</h2>
				</div>
				<div id="profile-interations">
					<button v-if="userId !== profileId" id="follow-button" @click="toggleFollow">
						{{ following ? "Unfollow" : "Follow" }}
					</button>
					<button v-if="userId !== profileId" id="ban-button" @click="toggleBan">{{
							banned ? "Unban" : "Ban"
						}}
					</button>
				</div>
			</div>
			<p>{{ profile.uploads }} photos | {{ profile.followers }} followers | {{ profile.following }} following</p>
			<hr style="border: 1px solid #4a4a4a; margin-bottom: 20px; width: 100%;"/>
			<p v-if="errormsg" id="error-msg">{{ errormsg }}</p>
			<div v-if="dataLoaded">
				<div v-for="(photo, index) in profile.photos" :key="index" style="margin-bottom: 15px">
					<PhotoCard @photoDeleted="this.fetchProfileData()"
							   :photoOwner="this.profileId"
							   :photoId="photo.photo_id"
							   :user="this.profileUsername"
							   :date="photo.date"
							   :likes="photo.total_likes"/>
				</div>
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
