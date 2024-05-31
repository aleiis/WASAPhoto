<script>

export default {
	props: ['photoOwner', 'photoId', 'user', 'date', 'likes'],
	emits: ['photoDeleted'],
	data() {
		return {
			currentUser: '',
			imgURL: null,
			liked: false,
			totalLikes: 0,
			totalComments: 0,
			commentSideHeight: 'auto',
			commentInput: '',
			comments: []
		}
	},
	methods: {

		/* utils */
		handleImageLoad() {
			const photoSideElement = this.$refs['photo-side']
			if (photoSideElement) {
				this.commentSideHeight = `${photoSideElement.clientHeight}px`
			}
		},

		limitBytes() {
			const bytes = new Blob([this.commentInput]).size;
			if (bytes > 128) {
				this.commentInput = this.commentInput.substring(0, this.commentInput.length - 1);
			}
		},

		/* image */
		async fetchImage() {
			try {
				let response = await this.$axios.get(`/users/${this.photoOwner}/photos/${this.photoId}/bin`, {
					responseType: 'blob',
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				this.imgURL = URL.createObjectURL(response.data)
			} catch (e) {
				console.error(e)
			}
		},

		/* comments */
		async loadComments() {
			try {
				let response = await this.$axios.get(`/users/${this.photoOwner}/photos/${this.photoId}/comments/`, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				this.comments = response.data.comments
				this.totalComments = this.comments.length
			} catch (e) {
				console.error(e)
			}
		},

		async postComment() {
			if (this.commentInput) {
				try {
					await this.$axios.post(`/users/${this.photoOwner}/photos/${this.photoId}/comments/`, {
							owner_id: parseInt(localStorage.getItem('userid')),
							content: this.commentInput
						},
						{
							headers: {
								Authorization: `Bearer ${localStorage.getItem('userid')}`
							}
						})
					await this.loadComments()
					this.commentInput = ''
				} catch (e) {
					console.error(e)
				}
			}
		},

		async deleteComment(commentId) {
			try {
				await this.$axios.delete(`/users/${this.photoOwner}/photos/${this.photoId}/comments/${commentId}`, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				await this.loadComments()
			} catch (e) {
				console.error(e)
			}
		},

		/* likes */
		async loadLike() {
			try {
				let response = await this.$axios.get(`/users/${this.photoOwner}/photos/${this.photoId}/likes/${localStorage.getItem('userid')}`, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				response.status === 200 ? this.liked = true : this.liked = false
			} catch (e) {
				if (e.response && e.response.status === 404) {
					this.liked = false
				} else {
					console.error(e)
				}
			}
		},

		toggleLike() {
			if (this.liked) {
				this.unlikePhoto()
			} else {
				this.likePhoto()
			}
		},

		async likePhoto() {
			try {
				await this.$axios.post(`/users/${this.photoOwner}/photos/${this.photoId}/likes/`, {
					liker: parseInt(localStorage.getItem('userid'), 10),
					photo: {
						owner_id: parseInt(this.photoOwner, 10),
						photo_id: parseInt(this.photoId, 10)
					}
				}, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				this.totalLikes++
				this.liked = true
			} catch (e) {
				console.error(e)
			}
		},

		async unlikePhoto() {
			try {
				await this.$axios.delete(`/users/${this.photoOwner}/photos/${this.photoId}/likes/${localStorage.getItem('userid')}`, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				this.totalLikes--
				this.liked = false
			} catch (e) {
				console.error(e)
			}
		},

		/* delete */
		async deletePhoto() {
			try {
				await this.$axios.delete(`/users/${this.photoOwner}/photos/${this.photoId}`, {
					headers: {
						Authorization: `Bearer ${localStorage.getItem('userid')}`
					}
				})
				this.$emit('photoDeleted')
			} catch (e) {
				console.error(e)
			}
		}
	},
	mounted() {
		this.currentUser = localStorage.getItem('username')
		this.totalLikes = this.likes
		this.fetchImage()
		this.loadComments()
		this.loadLike()
	}
};
</script>

<template>
	<div id="photo-card">
		<div ref="photo-side" id="photo-side">
			<img :src="imgURL" id="img" @load="handleImageLoad">
			<p style="margin-bottom: 0px;">{{ user }}, {{ date }}</p>
			<div style="display: flex; align-items: center;">
				<img v-if="liked" src='/src/assets/red_heart_icon.png' @click="toggleLike"
					 style="cursor: pointer; width: 30px;">
				<img v-else src='/src/assets/heart_icon.png' @click="toggleLike" style="cursor: pointer; width: 30px;">
				<p style="margin-left: 8px;">{{ totalLikes }} likes | {{ totalComments }} comments</p>
			</div>
			<button id="delete-foto" v-if="user === currentUser" @click="deletePhoto">Delete photo</button>
		</div>
		<div id="comments-side" :style="{ height: commentSideHeight }">
			<div id="post-comment"
				 style="margin-bottom: 10px; display: flex; flex-direction: row; align-content: baseline; min-width: 60%;">
				<input v-model="commentInput" @input="limitBytes" @keyup.enter="postComment" type="text"
					   placeholder="Add a comment..." id="comment-input">
				<button id="post-comment-button" :disabled="commentInput ? false : true" @click="postComment">Post
				</button>
			</div>
			<div v-for="(comment, index) in comments" :key="index"
				 style="width: 100%; display: flex; flex-direction: row; justify-content: space-between; align-content: center;">
				<p style="margin: 5px; margin-left: 10px">{{ comment.owner.username }}: {{ comment.content }}</p>
				<img v-if="comment.owner.username === currentUser" @click="deleteComment(comment.comment_id)"
					 src="\src\assets\red_delete_icon.png" alt="delete"
					 style="cursor: pointer; margin-top: 5px; height: 16px; margin-right: 10px; color: red;">
			</div>
		</div>
	</div>
</template>

<style scoped>

p {
	font-family: Arial, sans-serif;
	font-size: 14px;
}

#photo-card {
	background-color: #ffffff;
	border-radius: 10px;
	box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
	padding: 40px;
	display: flex;
	flex-direction: row;
	align-items: flex-start;
	width: calc(100% - 80px);
	height: fit-content;
}

#photo-side {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	max-width: 60%;
	max-height: 800px;
	min-width: 100px;
	height: auto;
	object-fit: cover;
}

#photo-side > img {
	background-color: #e7e7e7;
	border: 1px solid #ccc;
	border-radius: 4px;
	max-width: 100%;
	min-width: 160px;
	height: auto;
}

#delete-foto {
	background-color: rgba(0, 0, 0, 0);
	border: 1px solid red;
	border-radius: 4px;
	color: red;
	cursor: pointer;
	font-size: 12px;
	padding: 10px;
	margin-top: 5px;
	font-family: Arial, sans-serif;
	font-weight: bold;
}

#comments-side {
	background-color: #f5f5f5;
	border: 1px solid #ccc;
	border-radius: 4px;
	padding: 10px;
	margin-left: 20px;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	flex: 1;
	min-width: 40%;
	width: auto;
	overflow-y: auto;
}

#comment-input {
	border: 0px solid #ffffff;
	border-bottom: 1px solid #ccc;
	border-radius: 0px;
	font-size: 14px;
	padding: 10px;
	width: 100%;
}

#comment-input:focus {
	outline: none;
}

#post-comment-button {
	background-color: rgba(0, 0, 0, 0);
	border: 0px;
	border-radius: 4px;
	color: rgb(0, 153, 255);
	margin-top: 2px;
	cursor: pointer;
	font-size: 16px;
	padding: 10px;
	margin-left: 20px;
	font-family: Arial, sans-serif;
	font-weight: bold;
}

#post-comment-button:hover {
	background-color: rgba(0, 153, 255, 0.1);
	text-decoration: underline;
}

#post-comment-button:disabled {
	color: #ccc;
	cursor: not-allowed;
}

#post-comment-button:disabled:hover {
	background-color: rgba(0, 153, 255, 0);
}

</style>
